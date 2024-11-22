FROM php:8.2-apache

RUN a2enmod remoteip

RUN mkdir /var/www/adam-site

COPY php/ /var/www/adam-site/

COPY php/apache-extra.conf /etc/apache2/sites-available/000-default.conf
COPY php/apache2.conf /etc/apache2/apache2.conf
COPY php/php.ini /usr/local/etc/php/php.ini

CMD apache2-foreground