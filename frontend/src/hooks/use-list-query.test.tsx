import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, useLocation } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { useListQuery } from "./use-list-query";

function Probe() {
  const { limit, offset, q, setLimit, setOffset, setQuery } = useListQuery();
  const location = useLocation();

  return (
    <div>
      <output data-testid="state">{`${limit}|${offset}|${q}`}</output>
      <output data-testid="search">{location.search}</output>
      <button type="button" onClick={() => setOffset(40)}>
        offset 40
      </button>
      <button type="button" onClick={() => setLimit(50)}>
        limit 50
      </button>
      <button type="button" onClick={() => setQuery("api")}>
        query api
      </button>
      <button type="button" onClick={() => setOffset(0)}>
        first page
      </button>
    </div>
  );
}

function renderProbe(entry = "/projects") {
  render(
    <MemoryRouter initialEntries={[entry]}>
      <Probe />
    </MemoryRouter>,
  );
}

function state() {
  return screen.getByTestId("state").textContent;
}

describe("useListQuery", () => {
  it("falls back to the backend defaults", () => {
    renderProbe();

    expect(state()).toBe("20|0|");
  });

  it("reads limit, offset, and q from the url", () => {
    renderProbe("/projects?limit=10&offset=40&q=api");

    expect(state()).toBe("10|40|api");
  });

  it("ignores out of range values", () => {
    renderProbe("/projects?limit=500&offset=-5");

    expect(state()).toBe("20|0|");
  });

  it("writes the offset and drops it at zero", async () => {
    renderProbe();
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: "offset 40" }));
    expect(state()).toBe("20|40|");
    expect(screen.getByTestId("search")).toHaveTextContent("offset=40");

    await user.click(screen.getByRole("button", { name: "first page" }));
    expect(state()).toBe("20|0|");
    expect(screen.getByTestId("search")).not.toHaveTextContent("offset");
  });

  it("resets the offset when the page size changes", async () => {
    renderProbe("/projects?offset=60");
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: "limit 50" }));

    expect(state()).toBe("50|0|");
  });

  it("resets the offset when the query changes", async () => {
    renderProbe("/projects?offset=60&q=old");
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: "query api" }));

    expect(state()).toBe("20|0|api");
    expect(screen.getByTestId("search")).not.toHaveTextContent("offset");
  });
});
