import {
  Area,
  CartesianGrid,
  ComposedChart,
  Legend,
  Line,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'

import { formatNumber, formatShortDate } from '@/lib/utils'

export type TrendSeries = {
  key: string
  label: string
  color: string
  type?: 'area' | 'line'
  yAxis?: 'left' | 'right'
}

type TrendChartProps = {
  data: Array<Record<string, string | number>>
  series: TrendSeries[]
  height?: number
  dualAxis?: boolean
}

export function TrendChart({
  data,
  series,
  height = 280,
  dualAxis = false,
}: TrendChartProps) {
  return (
    <div style={{ height }} className="w-full">
      <ResponsiveContainer width="100%" height="100%">
        <ComposedChart
          data={data}
          margin={{ top: 8, right: 8, bottom: 0, left: -18 }}
        >
          <defs>
            {series.map((item) => (
              <linearGradient
                key={item.key}
                id={`gradient-${item.key}`}
                x1="0"
                y1="0"
                x2="0"
                y2="1"
              >
                <stop offset="0%" stopColor={item.color} stopOpacity={0.22} />
                <stop offset="100%" stopColor={item.color} stopOpacity={0} />
              </linearGradient>
            ))}
          </defs>
          <CartesianGrid
            vertical={false}
            strokeDasharray="3 3"
            className="stroke-slate-200 dark:stroke-slate-800"
          />
          <XAxis
            dataKey="date"
            tickFormatter={formatShortDate}
            tickLine={false}
            axisLine={false}
            minTickGap={28}
            tick={{ fontSize: 11, fill: 'currentColor' }}
            className="text-slate-400 dark:text-slate-500"
          />
          <YAxis
            yAxisId="left"
            tickLine={false}
            axisLine={false}
            width={48}
            tick={{ fontSize: 11, fill: 'currentColor' }}
            className="text-slate-400 dark:text-slate-500"
            tickFormatter={(value: number) => formatNumber(value)}
          />
          {dualAxis ? (
            <YAxis
              yAxisId="right"
              orientation="right"
              tickLine={false}
              axisLine={false}
              width={40}
              tick={{ fontSize: 11, fill: 'currentColor' }}
              className="text-slate-400 dark:text-slate-500"
              tickFormatter={(value: number) => formatNumber(value)}
            />
          ) : null}
          <Tooltip
            cursor={{ stroke: '#94a3b8', strokeDasharray: '4 4' }}
            labelFormatter={(label) => formatShortDate(String(label))}
            formatter={(value, name) => [
              formatNumber(Number(value)),
              series.find((item) => item.key === name)?.label ?? String(name),
            ]}
            contentStyle={{
              borderRadius: 12,
              border: '1px solid rgb(226 232 240)',
              boxShadow: '0 12px 32px rgb(15 23 42 / 0.08)',
              fontSize: 12,
            }}
          />
          <Legend
            iconType="circle"
            iconSize={7}
            wrapperStyle={{ fontSize: 12, paddingTop: 12 }}
          />
          {series.map((item) =>
            item.type === 'line' ? (
              <Line
                key={item.key}
                yAxisId={item.yAxis ?? 'left'}
                type="monotone"
                dataKey={item.key}
                name={item.label}
                stroke={item.color}
                strokeWidth={2}
                dot={false}
                activeDot={{ r: 4 }}
              />
            ) : (
              <Area
                key={item.key}
                yAxisId={item.yAxis ?? 'left'}
                type="monotone"
                dataKey={item.key}
                name={item.label}
                stroke={item.color}
                strokeWidth={2}
                fill={`url(#gradient-${item.key})`}
              />
            ),
          )}
        </ComposedChart>
      </ResponsiveContainer>
    </div>
  )
}
