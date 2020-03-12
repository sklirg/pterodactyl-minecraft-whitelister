Minecraft Whitelister for Pterodactyl
=====================================

A simple client using the Pterodactyl API for sending commands
to a minecraft server for whitelisting/un-whitelisting people.

Usage
-----

Send an HTTP POST request to ``/whitelist?username=<username>``,
where <username> is the username to whitelist.

Send an HTTP DELETE request to ``/whitelist?username=<username>``,
where <username> is the username to un-whitelist.

Installation
------------

The following environment variables need to be set:

- ``PT_API = Pterodactyl base URL (example: https://pterodactyl.app)``

- ``PT_SERVER_ID = Get this from the server URL when viewing a server, it is numeric.``

- ``PT_API_KEY = Creat an API key for a user with the permissions you want, e.g. restricted to a server.``

The following environment variables might be set:

- ``HOST = Sets the listening address``

- ``PORT = Sets the listening port``
