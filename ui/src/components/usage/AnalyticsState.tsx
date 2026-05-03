import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

interface AnalyticsStateProps {
  title: string;
  state: "empty" | "unavailable" | "invalid-filter";
  message?: string;
}

const defaultMessages: Record<AnalyticsStateProps["state"], string> = {
  empty: "No records matched the current filters.",
  unavailable: "Analytics are currently unavailable. Try again shortly.",
  "invalid-filter": "The selected filters are invalid. Adjust them and retry.",
};

export default function AnalyticsState({
  title,
  state,
  message,
}: AnalyticsStateProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent className="text-sm text-muted-foreground">
        {message || defaultMessages[state]}
      </CardContent>
    </Card>
  );
}
