require "test_helper"

class WeightTrendTest < ActiveSupport::TestCase
  Entry = Struct.new(:measured_on, :weight_kg, :body_fat_pct)

  def series(days, &blk) = (0...days).map { |i| Entry.new(Date.new(2026, 1, 1) + i, blk.call(i), nil) }

  test "smooths noise and measures loss" do
    t = WeightTrend.new(series(120) { |i| 100 - i * 0.1 + (i.even? ? 0.8 : -0.8) })
    assert_in_delta 100 - 119 * 0.1, t.current_trend, 1.5
    assert_operator t.total_loss_pct, :>, 9
    refute t.too_fast?
  end

  test "flags losing faster than 1.5% a week" do
    t = WeightTrend.new(series(90) { |i| 120 * (1 - 0.003)**i })
    assert t.too_fast?
  end

  test "flags regain" do
    t = WeightTrend.new(series(120) { |i| i < 60 ? 100 - i * 0.2 : 88 + (i - 60) * 0.12 })
    assert t.regaining?
  end

  test "plateau needs 12 weeks on medication and a flat trend" do
    t = WeightTrend.new(series(200) { |i| i < 120 ? 100 - i * 0.12 : 85.6 })
    assert t.plateau?(20)
    refute t.plateau?(8)
  end

  test "carries the trend across missing days" do
    entries = [Entry.new(Date.new(2026, 1, 1), 100, nil), Entry.new(Date.new(2026, 1, 15), 98, nil)]
    t = WeightTrend.new(entries)
    assert_equal 15, t.series.size
    assert_nil t.series[5][:w]
  end
end
