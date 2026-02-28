"use client";

import { useMemo } from "react";
import type { Subcategory } from "@/types/marketplace";

export function useCategoryColumns(items: Subcategory[], columnCount = 2) {
  return useMemo(() => {
    if (items.length === 0) {
      return [];
    }

    const size = Math.ceil(items.length / columnCount);
    const columns: Subcategory[][] = [];
    for (let index = 0; index < columnCount; index += 1) {
      const start = index * size;
      const end = start + size;
      const chunk = items.slice(start, end);
      if (chunk.length > 0) {
        columns.push(chunk);
      }
    }
    return columns;
  }, [items, columnCount]);
}
