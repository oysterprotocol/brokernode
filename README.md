# Brokernode

## Getting Started

The broker node uses Docker to spin up a go app, mysql, and private iota instance (TODO). You must first install [Docker](https://www.docker.com/community-edition).

```bash
# Starts the brokernode on port 3000
docker-compose up —-build # This takes a few minutes when you first run it.

# Executing commands in the app container
# Use `docker-compose exec YOUR_COMMAND`
# Eg: To run buffalo's test suite, run:
docker-compose exec app buffalo test

# Get a bash shell in the app container
docker-compose exec app bash

# Once in the app container, you can use all buffalo commands:
brokernode# buffalo db migrate
brokernode# buffalo test
```

---

# Welcome to Buffalo!

Thank you for choosing Buffalo for your web development needs.

## Database Setup

It looks like you chose to set up your application using a mysql database! Fantastic!

The first thing you need to do is open up the "database.yml" file and edit it to use the correct usernames, passwords, hosts, etc... that are appropriate for your environment.

You will also need to make sure that **you** start/install the database of your choice. Buffalo **won't** install and start mysql for you.

### Create Your Databases

Ok, so you've edited the "database.yml" file and started mysql, now Buffalo can create the databases in that file for you:

    $ buffalo db create -a

## Starting the Application

Buffalo ships with a command that will watch your application and automatically rebuild the Go binary and any assets for you. To do that run the "buffalo dev" command:

    $ buffalo dev

If you point your browser to [http://127.0.0.1:3000](http://127.0.0.1:3000) you should see a "Welcome to Buffalo!" page.

**Congratulations!** You now have your Buffalo application up and running.

## What Next?

We recommend you heading over to [http://gobuffalo.io](http://gobuffalo.io) and reviewing all of the great documentation there.

Good luck!

[Powered by Buffalo](http://gobuffalo.io)
