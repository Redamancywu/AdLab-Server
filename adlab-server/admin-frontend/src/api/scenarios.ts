import client, { unwrap } from './client'

export type ScenarioName =
  | 'high_fill_stable'
  | 'price_competition'
  | 'random_error'
  | 'no_fill'
  | 'high_latency'
  | 'mixed_behavior'

export const SCENARIO_META: Record<ScenarioName, { label: string; desc: string; color: string }> = {
  high_fill_stable:  { label: '高填充稳定', desc: '95% 填充率，固定出价 $1.5，低延迟低错误率，适合基准测试', color: 'green' },
  price_competition: { label: '价格竞争',   desc: '随机出价 $0.5~$3.0，80% 填充率，模拟多 DSP 激烈竞价', color: 'blue' },
  random_error:      { label: '随机错误',   desc: '30% 错误率（HTTP 500），70% 填充率，测试错误容忍', color: 'red' },
  no_fill:           { label: '无填充',     desc: '0% 填充率，测试 SDK 无广告时的降级处理', color: 'default' },
  high_latency:      { label: '高延迟',     desc: '300ms 基础延迟，5% 超时错误，测试超时处理', color: 'orange' },
  mixed_behavior:    { label: '混合行为',   desc: '概率出价三档，60% 填充，15% 错误，模拟真实复杂环境', color: 'purple' },
}

export const switchScenario = (scenario: ScenarioName) =>
  client
    .post('/admin/scenarios/switch', { scenario })
    .then(unwrap<{ scenario: string; updated_count: number }>)
