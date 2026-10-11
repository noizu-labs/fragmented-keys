import { defineConfig } from "tsup";

export default defineConfig({
  entry: [
    "src/configuration.ts",
    "src/interfaces/cache-handler.ts",
    "src/interfaces/cache-handler-sync.ts",
    "src/interfaces/tag.ts",
    "src/interfaces/key.ts",
    "src/interfaces/key-ring.ts",
    "src/interfaces/encryption.ts",
    "src/interfaces/types.ts",
    "src/cache-entry/cache-entry.ts",
    "src/cache-handler/memory-handler.ts",
    "src/cache-handler/redis-handler.ts",
    "src/tag/base-tag.ts",
    "src/tag/standard-tag.ts",
    "src/tag/constant-tag.ts",
    "src/key/standard-key.ts",
    "src/key-ring/fragmented-key-ring.ts",
  ],
  format: ["esm", "cjs"],
  dts: true,
  clean: true,
  splitting: false,
  sourcemap: true,
});
