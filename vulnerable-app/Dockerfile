FROM php:8.2-apache

# SQLite 확장 설치
RUN apt-get update && apt-get install -y \
    sqlite3 \
    libsqlite3-dev \
    && docker-php-ext-install pdo_sqlite \
    && rm -rf /var/lib/apt/lists/*

# 업로드 디렉토리 생성
RUN mkdir -p /tmp/uploads && chmod 777 /tmp/uploads

# Apache 설정: 파일 업로드 허용
RUN echo "file_uploads = On" >> /usr/local/etc/php/conf.d/uploads.ini && \
    echo "upload_max_filesize = 10M" >> /usr/local/etc/php/conf.d/uploads.ini && \
    echo "post_max_size = 10M" >> /usr/local/etc/php/conf.d/uploads.ini

# 웹 루트에 애플리케이션 복사
COPY index.php /var/www/html/

# Apache 포트 설정
EXPOSE 80

# Apache 시작
CMD ["apache2-foreground"]