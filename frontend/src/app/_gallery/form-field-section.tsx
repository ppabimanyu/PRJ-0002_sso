import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { FormField } from "@/components/shared/form-field";
import { Button } from "@/components/ui/button";
import { FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { GallerySection } from "./section";

const formSchema = z.object({
  name: z.string().min(2, "Enter at least 2 characters."),
  message: z.string().min(10, "Enter at least 10 characters."),
});

type FormValues = z.infer<typeof formSchema>;

export function FormFieldSection() {
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: { name: "", message: "" },
  });

  function handleSubmit(values: FormValues) {
    toast.success("Form submitted", { description: `Thanks, ${values.name}.` });
    form.reset();
  }

  return (
    <GallerySection
      title="Form field"
      description="React Hook Form and Zod, with label, description, and error wiring."
    >
      <form className="flex max-w-md flex-col gap-6" onSubmit={form.handleSubmit(handleSubmit)}>
        <FieldGroup>
          <FormField
            control={form.control}
            name="name"
            label="Name"
            description="Shown on the submitted message."
            render={({ field, id, describedBy, invalid }) => (
              <Input
                {...field}
                id={id}
                aria-describedby={describedBy}
                aria-invalid={invalid}
                placeholder="Ada Lovelace"
              />
            )}
          />
          <FormField
            control={form.control}
            name="message"
            label="Message"
            description="At least 10 characters."
            render={({ field, id, describedBy, invalid }) => (
              <Textarea
                {...field}
                id={id}
                aria-describedby={describedBy}
                aria-invalid={invalid}
                placeholder="What should we build next?"
              />
            )}
          />
        </FieldGroup>
        <div className="flex items-center gap-2">
          <Button type="submit">Submit</Button>
          <Button type="button" variant="ghost" onClick={() => form.reset()}>
            Reset
          </Button>
        </div>
      </form>
    </GallerySection>
  );
}
