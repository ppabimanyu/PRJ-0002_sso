import { useId, type ReactNode } from "react";
import {
  Controller,
  type Control,
  type ControllerFieldState,
  type ControllerRenderProps,
  type FieldPath,
  type FieldValues,
  type FormState,
} from "react-hook-form";

import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import { cn } from "@/lib/utils";

type FormFieldRenderProps<
  TFieldValues extends FieldValues,
  TName extends FieldPath<TFieldValues>,
> = {
  field: ControllerRenderProps<TFieldValues, TName>;
  fieldState: ControllerFieldState;
  formState: FormState<TFieldValues>;
  id: string;
  describedBy: string | undefined;
  invalid: boolean;
};

type FormFieldProps<TFieldValues extends FieldValues, TName extends FieldPath<TFieldValues>> = {
  control: Control<TFieldValues>;
  name: TName;
  label: ReactNode;
  description?: ReactNode;
  className?: string;
  render: (props: FormFieldRenderProps<TFieldValues, TName>) => ReactNode;
};

export function FormField<TFieldValues extends FieldValues, TName extends FieldPath<TFieldValues>>({
  control,
  name,
  label,
  description,
  className,
  render,
}: FormFieldProps<TFieldValues, TName>) {
  const id = useId();
  const descriptionId = `${id}-description`;
  const errorId = `${id}-error`;

  return (
    <Controller
      control={control}
      name={name}
      render={({ field, fieldState, formState }) => {
        const invalid = fieldState.invalid;
        const describedBy =
          [description ? descriptionId : null, invalid && fieldState.error ? errorId : null]
            .filter(Boolean)
            .join(" ") || undefined;

        return (
          <Field data-invalid={invalid ? "true" : undefined} className={cn(className)}>
            <FieldLabel htmlFor={id}>{label}</FieldLabel>
            {render({ field, fieldState, formState, id, describedBy, invalid })}
            {description ? (
              <FieldDescription id={descriptionId}>{description}</FieldDescription>
            ) : null}
            {invalid ? <FieldError id={errorId} errors={[fieldState.error]} /> : null}
          </Field>
        );
      }}
    />
  );
}
