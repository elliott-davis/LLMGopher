type ModelRateLimitStatusProps = {
  rateLimitRPS?: number | null;
};

export default function ModelRateLimitStatus({
  rateLimitRPS,
}: ModelRateLimitStatusProps) {
  if (typeof rateLimitRPS === "number" && rateLimitRPS > 0) {
    return (
      <span className="font-medium">
        {rateLimitRPS} requests/sec
      </span>
    );
  }

  return (
    <span className="text-muted-foreground">
      No model-level limit
    </span>
  );
}
