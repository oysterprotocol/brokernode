
## Base Image is vanilla php and apache

FROM php:7.0-apache

## Get the latest webnode source code from GitHub

RUN git clone https://github.com/oysterprotocol/brokernode.git /var/www/html/

## Expose port 80 for web

EXPOSE 80

