interface ProviderLogoProps {
  name: string;
  size?: "lg" | "";
}

export function ProviderLogo({ name, size = "" }: ProviderLogoProps) {
  const lower = name.toLowerCase();
  let kind = "openai";
  let letter = "O";
  if (lower.includes("anthropic") || lower.includes("claude")) {
    kind = "anthropic"; letter = "A";
  } else if (
    lower.includes("vertex") || lower.includes("google") || lower.includes("gemini")
  ) {
    kind = "google"; letter = "G";
  } else if (lower.includes("mistral")) {
    kind = "mistral"; letter = "M";
  } else if (lower.includes("bedrock") || lower.includes("aws")) {
    kind = "bedrock"; letter = "B";
  } else if (lower.includes("vllm") || lower.includes("vl")) {
    kind = "vllm"; letter = "v";
  }
  return (
    <span className={["logo", kind, size].filter(Boolean).join(" ")}>{letter}</span>
  );
}
