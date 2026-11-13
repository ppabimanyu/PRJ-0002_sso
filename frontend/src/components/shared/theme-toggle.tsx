import { Monitor, Moon, Sun } from "lucide-react";

import { useTheme, type Theme } from "@/app/providers";
import { Button } from "@/components/ui/button";

const themeOrder: Theme[] = ["light", "dark", "system"];

const themeLabels: Record<Theme, string> = {
  light: "Light",
  dark: "Dark",
  system: "System",
};

const themeIcons: Record<Theme, typeof Sun> = {
  light: Sun,
  dark: Moon,
  system: Monitor,
};

export function ThemeToggle() {
  const { theme, setTheme } = useTheme();
  const nextTheme = themeOrder[(themeOrder.indexOf(theme) + 1) % themeOrder.length];
  const Icon = themeIcons[theme];

  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={() => setTheme(nextTheme)}
      aria-label={`${themeLabels[theme]} theme. Switch to ${themeLabels[nextTheme]}.`}
      title={`${themeLabels[theme]} theme`}
    >
      <Icon />
    </Button>
  );
}
