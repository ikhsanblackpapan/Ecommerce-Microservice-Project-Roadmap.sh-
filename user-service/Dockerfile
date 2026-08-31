# Gunakan image resmi Node.js versi 22 alpine (ringan)
FROM node:22-alpine

# Tentukan folder kerja di dalam kontainer
WORKDIR /app

# Copy package.json dan package-lock.json untuk install dependensi
COPY package*.json ./

# Install dependensi aplikasi
RUN npm install

# Copy seluruh kode aplikasi ke dalam kontainer
COPY . .

# Generate Prisma Client di dalam kontainer
RUN npx prisma generate

# Buka port 8001 untuk diakses
EXPOSE 8001

# Jalankan aplikasi
CMD ["node", "server.js"]