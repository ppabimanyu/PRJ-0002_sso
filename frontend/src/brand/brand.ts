// Every project in the house renames itself here and nowhere else.
export const appName = import.meta.env.VITE_APP_NAME || "SSO";

/** Cells in the house test chart. The grid is fixed across every service. */
export const chartCells = 36;

/**
 * Which cell a service lights on the house chart. Deterministic, so the same
 * name always lands on the same cell and the graphic survives a reload or a
 * rename back and forth.
 */
export function serviceCell(name: string): number {
  let hash = 0;

  for (const character of name) {
    hash = (hash * 31 + character.charCodeAt(0)) >>> 0;
  }

  return hash % chartCells;
}
