# Gunakan image resmi PHP 8.2 dengan FPM berbasis Alpine Linux
FROM php:8.2-fpm-alpine

# Install ekstensi sistem dan dependensi yang dibutuhkan Laravel

# Install ekstensi sistem dan dependensi yang dibutuhkan Laravel
RUN apk add --no-cache \
    nginx \
    supervisor \
    curl \
    libpng-dev \
    libxml2-dev \
    zip \
    unzip \
    git

    # Install ekstensi PHP untuk MySQL
RUN docker-php-ext-install pdo_mysql bcmath

# Ambil Composer resmi dari image composer terpisah
COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

# Tentukan folder kerja di dalam kontainer
WORKDIR /var/www

# Copy seluruh kode Laravel ke dalam kontainer
COPY . .

# Install dependensi vendor Laravel via Composer (optimasi untuk production)
RUN composer install --no-interaction --optimize-autoloader --no-dev

# Atur hak akses folder storage dan bootstrap cache agar bisa ditulis oleh Laravel
RUN chown -R www-data:www-data /var/www/storage /var/www/bootstrap/cache

# Buka port 8002 untuk diakses luar kontainer
EXPOSE 8002

# Jalankan server internal PHP untuk melayani request
CMD ["php", "artisan", "serve", "--host=0.0.0.0", "--port=8002"]

