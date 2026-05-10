interface SparklineProps {
  data: number[];
  width?: number;
  height?: number;
  color?: string;
}

export function Sparkline({
  data,
  width = 160,
  height = 40,
  color = "var(--brand)",
}: SparklineProps) {
  if (!data || data.length < 2) return null;
  const max = Math.max(...data);
  const min = Math.min(...data);
  const range = Math.max(max - min, 1e-3);
  const pts = data.map((v, i) => {
    const x = (i / (data.length - 1)) * width;
    const y = height - ((v - min) / range) * (height - 4) - 2;
    return [x, y] as const;
  });
  const linePath =
    "M" + pts.map(([x, y]) => `${x.toFixed(1)},${y.toFixed(1)}`).join(" L ");
  const areaPath =
    `M${pts[0][0].toFixed(1)},${height} L ` +
    pts.map(([x, y]) => `${x.toFixed(1)},${y.toFixed(1)}`).join(" L ") +
    ` L ${pts[pts.length - 1][0].toFixed(1)},${height} Z`;

  return (
    <svg
      width={width}
      height={height}
      viewBox={`0 0 ${width} ${height}`}
      preserveAspectRatio="none"
      aria-hidden
      style={{ display: "block", overflow: "visible" }}
    >
      <path
        d={areaPath}
        fill={`color-mix(in oklch, ${color} 18%, transparent)`}
      />
      <path
        d={linePath}
        fill="none"
        stroke={color}
        strokeWidth="1.5"
        strokeLinejoin="round"
        strokeLinecap="round"
      />
    </svg>
  );
}
