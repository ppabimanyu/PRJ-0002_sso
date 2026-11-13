import { act, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { SearchInput } from "./search-input";

function setup(onValueChange = vi.fn()) {
  render(<SearchInput value="" onValueChange={onValueChange} debounceMs={300} />);

  return { onValueChange, input: screen.getByRole("textbox", { name: "Search" }) };
}

afterEach(() => {
  vi.useRealTimers();
});

describe("SearchInput", () => {
  it("waits for the debounce before reporting keystrokes", () => {
    vi.useFakeTimers();
    const { onValueChange, input } = setup();

    fireEvent.change(input, { target: { value: "api" } });

    expect(onValueChange).not.toHaveBeenCalled();

    act(() => {
      vi.advanceTimersByTime(299);
    });
    expect(onValueChange).not.toHaveBeenCalled();

    act(() => {
      vi.advanceTimersByTime(1);
    });
    expect(onValueChange).toHaveBeenCalledTimes(1);
    expect(onValueChange).toHaveBeenCalledWith("api");
  });

  it("reports the clear action immediately", () => {
    vi.useFakeTimers();
    const { onValueChange, input } = setup();

    fireEvent.change(input, { target: { value: "api" } });
    act(() => {
      vi.advanceTimersByTime(300);
    });
    onValueChange.mockClear();

    fireEvent.click(screen.getByRole("button", { name: "Clear search" }));

    expect(onValueChange).toHaveBeenCalledWith("");
    expect(input).toHaveValue("");
  });

  it("follows an external value change", () => {
    const { rerender } = render(<SearchInput value="first" onValueChange={vi.fn()} />);

    expect(screen.getByRole("textbox", { name: "Search" })).toHaveValue("first");

    rerender(<SearchInput value="second" onValueChange={vi.fn()} />);

    expect(screen.getByRole("textbox", { name: "Search" })).toHaveValue("second");
  });
});
