# Smooths noisy scale readings into a trend (exponential moving average, carried across
# missing days), then reads the signals that matter clinically: pace of loss, total loss,
# plateaus and regain.
class WeightTrend
  ALPHA = 0.1
  Point = Data.define(:date, :weight, :trend)

  attr_reader :points

  def initialize(entries)
    @entries = entries.sort_by(&:measured_on)
    @points = build
  end

  def empty? = @points.empty?
  def start_weight = @entries.first&.weight_kg&.to_f
  def current_trend = @points.last&.trend
  def latest_weight = @entries.last&.weight_kg&.to_f

  def total_loss_kg
    return 0.0 if empty?
    (start_weight - current_trend).round(1)
  end

  def total_loss_pct
    return 0.0 if empty? || start_weight.zero?
    (100.0 * total_loss_kg / start_weight).round(1)
  end

  # Percent of body weight lost per week over the last `days` days of the trend (positive = losing).
  def weekly_rate_pct(days = 28)
    return nil if @points.size < 2
    last = @points.last
    first = @points.reverse.find { |p| (last.date - p.date) >= days } || @points.first
    span = (last.date - first.date).to_f
    return nil if span < 7
    (100.0 * (first.trend - last.trend) / first.trend / (span / 7)).round(2)
  end

  def too_fast? = (r = weekly_rate_pct) && r > 1.5
  def plateau?(weeks_on_medication)
    r = weekly_rate_pct
    weeks_on_medication.to_i >= 12 && r && r.abs < 0.15
  end

  # Regain: trend up more than 2% over four weeks.
  def regaining?
    r = weekly_rate_pct(28)
    r && r < -0.5
  end

  def series
    @points.map { |p| { d: p.date.iso8601, w: p.weight&.round(2), t: p.trend.round(2) } }
  end

  private

  def build
    return [] if @entries.empty?
    by_date = @entries.index_by(&:measured_on)
    trend = @entries.first.weight_kg.to_f
    (by_date.keys.min..by_date.keys.max).map do |d|
      e = by_date[d]
      trend += ALPHA * (e.weight_kg.to_f - trend) if e
      Point.new(date: d, weight: e&.weight_kg&.to_f, trend: trend)
    end
  end
end
