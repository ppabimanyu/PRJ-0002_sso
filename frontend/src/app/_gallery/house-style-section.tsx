import { Activity, Database, Monitor, Server } from "lucide-react";

import { BrandLockup, BrandMark } from "@/brand/brand-mark";
import { ServiceChart } from "@/brand/service-chart";
import { cn } from "@/lib/utils";
import { GalleryDemo, GallerySection } from "./section";

const palette = [
  { role: "Brand", className: "bg-brand", token: "--brand" },
  { role: "Brand soft", className: "bg-brand-soft", token: "--brand-soft" },
  { role: "Ink", className: "bg-foreground", token: "--foreground" },
  { role: "Paper", className: "bg-background", token: "--background" },
  { role: "Muted text", className: "bg-muted-foreground", token: "--muted-foreground" },
  { role: "Line", className: "bg-border", token: "--border" },
  { role: "Success", className: "bg-success", token: "--success" },
  { role: "Warning", className: "bg-warning", token: "--warning" },
  { role: "Destructive", className: "bg-destructive", token: "--destructive" },
];

const markSizes = [
  { size: "size-4", label: "16" },
  { size: "size-6", label: "24" },
  { size: "size-9", label: "36" },
  { size: "size-14", label: "56" },
];

export function HouseStyleSection() {
  return (
    <GallerySection
      title="House style"
      description="The identity every service in the house inherits: one mark, one brand ink, one typeface, and colour reserved for meaning."
    >
      <GalleryDemo label="Mark and lockup">
        <div className="flex flex-wrap items-end gap-8">
          {markSizes.map(({ size, label }) => (
            <div key={label} className="flex flex-col items-center gap-2">
              <BrandMark className={cn(size, "text-brand")} />
              <span className="text-xs text-muted-foreground">{label}px</span>
            </div>
          ))}
        </div>
        <BrandLockup />
      </GalleryDemo>

      <GalleryDemo label="Service chart">
        <div className="flex flex-wrap items-start gap-8">
          {["Billing", "Console", "Ledger"].map((name) => (
            <div key={name} className="flex flex-col items-center gap-2">
              <ServiceChart name={name} className="w-40" />
              <span className="text-xs text-muted-foreground">{name}</span>
            </div>
          ))}
        </div>
        <p className="max-w-[60ch] text-sm text-muted-foreground">
          One fixed chart, one lit cell per service. The cell comes from the service name, so every
          project carries the same graphic with its own fingerprint.
        </p>
      </GalleryDemo>

      <GalleryDemo label="Palette">
        <div className="grid gap-4 sm:grid-cols-3">
          {palette.map(({ role, className, token }) => (
            <div key={role} className="flex items-center gap-3">
              <span
                className={cn("size-9 shrink-0 rounded-md border border-border", className)}
                aria-hidden="true"
              />
              <div>
                <p className="text-sm">{role}</p>
                <p className="font-mono text-xs text-muted-foreground">{token}</p>
              </div>
            </div>
          ))}
        </div>
      </GalleryDemo>

      <GalleryDemo label="Type scale">
        <div className="flex flex-col gap-4">
          <p className="text-[clamp(2rem,4vw,2.75rem)] leading-[1.05] font-semibold tracking-[-0.03em]">
            Display
          </p>
          <p className="text-2xl font-semibold tracking-tight">Headline</p>
          <p className="max-w-[60ch] text-base leading-relaxed">
            Body copy sets the reading rhythm at 16px with a 1.6 line height and no more than 60
            characters per line.
          </p>
          <p className="text-sm text-muted-foreground">Label and supporting text</p>
        </div>
      </GalleryDemo>

      <GalleryDemo label="Icons">
        <div className="flex flex-wrap items-center gap-6 text-muted-foreground">
          <Server className="size-5" aria-hidden="true" />
          <Monitor className="size-5" aria-hidden="true" />
          <Database className="size-5" aria-hidden="true" />
          <Activity className="size-5" aria-hidden="true" />
          <BrandMark className="size-5 text-brand" />
        </div>
        <p className="max-w-[60ch] text-sm text-muted-foreground">
          Lucide at the preset's stroke. The house layer tints with `--brand`, which maps to
          `--primary`, so a preset swap carries the graphics along.
        </p>
      </GalleryDemo>
    </GallerySection>
  );
}
