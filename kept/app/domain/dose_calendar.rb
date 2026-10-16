# Builds an iCalendar feed: one recurring event for the dosing schedule (weekly on the dose
# day, or daily for pills) with a reminder, plus a one-off "talk to your prescriber" event when
# the label's escalation window is reached. Lines are folded and text escaped per RFC 5545.
class DoseCalendar
  def initialize(user, now: Time.current)
    @user = user
    @med = user.medication
    @now = now
  end

  def to_ics
    lines = ["BEGIN:VCALENDAR", "VERSION:2.0", "PRODID:-//Kept//Dose reminders//EN", "CALSCALE:GREGORIAN",
             "METHOD:PUBLISH", "X-WR-CALNAME:Kept dose reminders", "X-PUBLISHED-TTL:PT12H", "REFRESH-INTERVAL;VALUE=DURATION:PT12H"]
    lines.concat(dose_event) if @med
    lines.concat(review_event) if @med && review_on
    lines << "END:VCALENDAR"
    lines.map { |l| fold(l) }.join("\r\n") + "\r\n"
  end

  private

  def drug = DrugCatalog.fetch(@med.drug)
  def zone = ActiveSupport::TimeZone[@user.time_zone] || Time.zone
  def stamp(t) = t.utc.strftime("%Y%m%dT%H%M%SZ")

  def first_dose_at
    d = @med.started_on
    d += 1 while drug.frequency == :weekly && @med.dose_weekday && d.wday != @med.dose_weekday
    zone.local(d.year, d.month, d.day, 9)
  end

  def dose_event
    rule = drug.frequency == :weekly ? "RRULE:FREQ=WEEKLY" : "RRULE:FREQ=DAILY"
    what = drug.form == :pill ? "Take #{@med.label}" : "#{@med.label} injection"
    ["BEGIN:VEVENT", "UID:dose-#{@med.id}@kept.app", "DTSTAMP:#{stamp(@now)}", "DTSTART;TZID=#{zone.tzinfo.name}:#{first_dose_at.strftime('%Y%m%dT%H%M%S')}",
     "DURATION:PT15M", rule, "SUMMARY:#{esc(what)}", "DESCRIPTION:#{esc("#{drug.how_to_take} Log it in Kept so your schedule and side-effect pattern stay accurate.")}",
     "BEGIN:VALARM", "ACTION:DISPLAY", "DESCRIPTION:#{esc(what)}", "TRIGGER:PT0M", "END:VALARM", "END:VEVENT"]
  end

  # When the label's minimum time at this dose is up, a nudge to talk with the prescriber.
  def review_on
    s = DoseSchedule.new(@med, [], now: @now)
    return nil if s.max_dose? || s.next_step.nil?
    d = @med.dose_started_on + drug.step_weeks * 7
    d >= @now.to_date ? d : nil
  end

  def review_event
    ["BEGIN:VEVENT", "UID:review-#{@med.id}-#{review_on.iso8601}@kept.app", "DTSTAMP:#{stamp(@now)}",
     "DTSTART;VALUE=DATE:#{review_on.strftime('%Y%m%d')}", "SUMMARY:#{esc("Dose check-in with your prescriber (#{drug.brand})")}",
     "DESCRIPTION:#{esc("You'll have been on #{@med.label} for #{drug.step_weeks} weeks — the label's minimum before the next step. Bring your Kept report.")}",
     "END:VEVENT"]
  end

  def esc(text) = text.to_s.gsub("\\", "\\\\\\\\").gsub(";", "\;").gsub(",", "\\,").gsub("\n", "\\n")

  # RFC 5545: lines longer than 75 octets continue on the next line after a space.
  def fold(line)
    return line if line.bytesize <= 75
    out, cur = [], +""
    line.each_char do |ch|
      if (cur + ch).bytesize > (out.empty? ? 75 : 74)
        out << cur
        cur = +""
      end
      cur << ch
    end
    out << cur
    out.first + out.drop(1).map { |c| "\r\n #{c}" }.join
  end
end
