import "dotenv/config";
import { defineConfig, env } from "prisma/config";

export default defineConfig({
  // Menentukan lokasi schema prisma
  schema: "prisma/schema.prisma", 
  
  // Mengatur folder output migrasi
  migrations: {
    path: "prisma/migrations",
  },
  
  // Konfigurasi koneksi database untuk CLI
  datasource: {
    url: env("DATABASE_URL"),
  },
});
