import type { ReactNode } from "react";

type GallerySectionProps = {
  title: string;
  description: string;
  children: ReactNode;
};

export function GallerySection({ title, description, children }: GallerySectionProps) {
  return (
    <section className="flex flex-col gap-4">
      <div className="flex flex-col gap-1">
        <h2 className="text-lg font-medium">{title}</h2>
        <p className="text-sm text-muted-foreground">{description}</p>
      </div>
      {children}
    </section>
  );
}

export function GalleryDemo({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="flex flex-col gap-3">
      <h3 className="text-sm font-medium">{label}</h3>
      {children}
    </div>
  );
}
