import { StatusBadge } from "@/components/shared/status-badge";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { GalleryDemo, GallerySection } from "./section";

const badgeVariants = ["default", "secondary", "destructive", "outline"] as const;

const rows = [
  { name: "api-gateway", owner: "Ada Lovelace", updated: "2 minutes ago" },
  { name: "billing-worker", owner: "Grace Hopper", updated: "1 hour ago" },
  { name: "web-console", owner: "Alan Turing", updated: "yesterday" },
];

export function DisplaySection() {
  return (
    <GallerySection
      title="Display"
      description="Badge, skeleton, tabs, avatar, and the table markup that the data table composition builds on."
    >
      <GalleryDemo label="Badge">
        <div className="flex flex-wrap items-center gap-2">
          {badgeVariants.map((variant) => (
            <Badge key={variant} variant={variant}>
              {variant}
            </Badge>
          ))}
        </div>
      </GalleryDemo>

      <GalleryDemo label="Status badge">
        <div className="flex flex-wrap items-center gap-2">
          <StatusBadge status="success" label="Operational" />
          <StatusBadge status="warning" label="Degraded" />
          <StatusBadge status="danger" label="Failed" />
          <StatusBadge status="neutral" label="Draft" />
          <StatusBadge status="pending" label="Checking" />
          <StatusBadge status="neutral" label="No dot" dot={false} />
        </div>
      </GalleryDemo>

      <GalleryDemo label="Skeleton">
        <div className="flex max-w-md flex-col gap-2">
          <Skeleton className="h-4 w-48" />
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-3/4" />
        </div>
      </GalleryDemo>

      <GalleryDemo label="Tabs">
        <Tabs defaultValue="general" className="max-w-md">
          <TabsList>
            <TabsTrigger value="general">General</TabsTrigger>
            <TabsTrigger value="members">Members</TabsTrigger>
          </TabsList>
          <TabsContent value="general" className="text-sm text-muted-foreground">
            Peer sections of the same page, not routes.
          </TabsContent>
          <TabsContent value="members" className="text-sm text-muted-foreground">
            Each tab keeps its own panel content.
          </TabsContent>
        </Tabs>
      </GalleryDemo>

      <GalleryDemo label="Avatar">
        <div className="flex items-center gap-2">
          <Avatar>
            <AvatarFallback>AL</AvatarFallback>
          </Avatar>
          <Avatar>
            <AvatarFallback>GH</AvatarFallback>
          </Avatar>
          <Avatar>
            <AvatarFallback>AT</AvatarFallback>
          </Avatar>
        </div>
      </GalleryDemo>

      <GalleryDemo label="Table">
        <Table>
          <TableCaption>Markup only. Behaviour belongs to the data table composition.</TableCaption>
          <TableHeader>
            <TableRow>
              <TableHead>Project</TableHead>
              <TableHead>Owner</TableHead>
              <TableHead>Updated</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((row) => (
              <TableRow key={row.name}>
                <TableCell className="font-medium">{row.name}</TableCell>
                <TableCell>{row.owner}</TableCell>
                <TableCell className="text-muted-foreground">{row.updated}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </GalleryDemo>
    </GallerySection>
  );
}
