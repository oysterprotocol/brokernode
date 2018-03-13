Oyster Web Storage
Broker Node README

## Installation (Mac)

Install PHP v7.0+ and composer using Homebrew (tutorial here: https://gist.github.com/shashankmehta/6ff13acd60f449eea6311cba4aae900a)

    xcode-select --install
    brew update
    brew tap homebrew/dupes
    brew tap homebrew/php
    brew install php70
    brew install mcrypt php70-mcrypt
    brew install composer
    export PATH="$(brew --prefix homebrew/php/php70)/bin:$PATH"
    php --version

Install Docker
https://docs.docker.com/docker-for-mac/

Go to the broker-node subfolder, install the dependencies, run migrations on the database, and then start the server

    composer install // installs dependencies
    php artisan migrate // run database migrations

*Note: for me I had to update DB_HOST=127.0.0.1 in broker-node/.env to run the migration command successfully. Change it back to DB_HOST=mariadb after you do this.*

    composer start // start the server

You can now make requests to the server at: http://localhost:8000
