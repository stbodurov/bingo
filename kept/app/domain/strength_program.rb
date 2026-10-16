# A three-day, full-body resistance program built to preserve muscle during GLP-1 weight
# loss: big movement patterns, 2–3 hard sets, progressive load. Home and gym variants.
module StrengthProgram
  Exercise = Data.define(:name, :home, :sets, :reps, :cue)
  Day = Data.define(:key, :title, :exercises)

  DAYS = [
    Day.new(key: "A", title: "Day A — squat & push", exercises: [
      Exercise.new(name: "Goblet squat", home: "Chair squat with a backpack", sets: 3, reps: "8–12", cue: "Sit back, chest tall, full depth you control."),
      Exercise.new(name: "Dumbbell bench press", home: "Push-ups (hands on a counter if needed)", sets: 3, reps: "8–12", cue: "Lower slowly, press hard."),
      Exercise.new(name: "Seated cable row", home: "Backpack bent-over row", sets: 3, reps: "10–12", cue: "Squeeze shoulder blades together."),
      Exercise.new(name: "Farmer carry", home: "Carry two grocery bags", sets: 2, reps: "40 m", cue: "Tall posture, tight grip.")]),
    Day.new(key: "B", title: "Day B — hinge & pull", exercises: [
      Exercise.new(name: "Romanian deadlift", home: "Single-leg hip hinge", sets: 3, reps: "8–10", cue: "Hips back, flat back, feel the hamstrings."),
      Exercise.new(name: "Lat pulldown", home: "Band pulldown or doorframe row", sets: 3, reps: "10–12", cue: "Pull elbows to your ribs."),
      Exercise.new(name: "Overhead press", home: "Pike push-up or water-jug press", sets: 3, reps: "8–12", cue: "Ribs down, press straight up."),
      Exercise.new(name: "Plank", home: "Plank", sets: 3, reps: "30–45 s", cue: "Squeeze glutes, breathe.")]),
    Day.new(key: "C", title: "Day C — single leg & full body", exercises: [
      Exercise.new(name: "Split squat", home: "Split squat", sets: 3, reps: "8–10 / leg", cue: "Front knee tracks over toes."),
      Exercise.new(name: "Incline dumbbell press", home: "Incline push-up", sets: 3, reps: "10–12", cue: "Control the lowering."),
      Exercise.new(name: "Hip thrust", home: "Glute bridge", sets: 3, reps: "10–15", cue: "Pause at the top."),
      Exercise.new(name: "Step-ups", home: "Stair step-ups", sets: 2, reps: "10 / leg", cue: "Drive through the whole foot.")])
  ].freeze

  def self.next_day(workouts)
    last = workouts.select { |w| w.kind == "strength" && w.program_day.present? }.max_by(&:performed_on)
    return DAYS.first unless last
    DAYS[(DAYS.index { |d| d.key == last.program_day }.to_i + 1) % DAYS.size]
  end
end
