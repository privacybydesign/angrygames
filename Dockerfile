FROM php:8.0-apache

WORKDIR /var/www/html

ADD . .

RUN chown -R www-data:www-data /var/www/html \
    && chmod -R 755 /var/www/html

RUN echo Listen 8080 > /etc/apache2/ports.conf 

EXPOSE 8080

