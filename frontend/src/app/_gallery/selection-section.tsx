import { ChevronDownIcon, PencilIcon, TrashIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { GalleryDemo, GallerySection } from "./section";

const pageSizes = ["10", "20", "50", "100"];

export function SelectionSection() {
  return (
    <GallerySection
      title="Selection"
      description="Select for form values, dropdown menu for actions. They are not interchangeable."
    >
      <div className="grid gap-8 md:grid-cols-2">
        <GalleryDemo label="Select">
          <Select defaultValue="20">
            <SelectTrigger className="w-56" aria-label="Rows per page">
              <SelectValue placeholder="Rows per page" />
            </SelectTrigger>
            <SelectContent>
              {pageSizes.map((size) => (
                <SelectItem key={size} value={size}>
                  {size} rows
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </GalleryDemo>

        <GalleryDemo label="Dropdown menu">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" className="w-56 justify-between">
                Row actions
                <ChevronDownIcon />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start">
              <DropdownMenuLabel>Project</DropdownMenuLabel>
              <DropdownMenuItem>
                <PencilIcon />
                Rename
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem variant="destructive">
                <TrashIcon />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </GalleryDemo>
      </div>
    </GallerySection>
  );
}
