import { zodResolver } from "@hookform/resolvers/zod";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useForm } from "react-hook-form";
import { describe, expect, it, vi } from "vitest";
import { z } from "zod";

import { Input } from "@/components/ui/input";
import { FormField } from "./form-field";

const formSchema = z.object({
  name: z.string().min(2, "Enter at least 2 characters."),
});

type FormValues = z.infer<typeof formSchema>;

function TestForm({ onSubmit }: { onSubmit: (values: FormValues) => void }) {
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: { name: "" },
  });

  return (
    <form onSubmit={form.handleSubmit(onSubmit)}>
      <FormField
        control={form.control}
        name="name"
        label="Name"
        description="Your display name."
        render={({ field, id, describedBy, invalid }) => (
          <Input {...field} id={id} aria-describedby={describedBy} aria-invalid={invalid} />
        )}
      />
      <button type="submit">Save</button>
    </form>
  );
}

describe("FormField", () => {
  it("associates the label, description, and error with the control", async () => {
    const onSubmit = vi.fn();

    render(<TestForm onSubmit={onSubmit} />);

    const input = screen.getByRole("textbox", {
      name: "Name",
      description: "Your display name.",
    });

    await userEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByText("Enter at least 2 characters.")).toBeInTheDocument();
    expect(input).toHaveAttribute("aria-invalid", "true");
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("submits a valid value", async () => {
    const onSubmit = vi.fn();

    render(<TestForm onSubmit={onSubmit} />);

    await userEvent.type(screen.getByRole("textbox", { name: "Name" }), "Ada");
    await userEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(onSubmit).toHaveBeenCalledWith({ name: "Ada" }, expect.anything());
  });
});
