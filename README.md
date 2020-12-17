# Swarmonitor

Periodically checks if there is at least one container running for each configured container set and sends e-mails to notify when no container is running. It uses the Docker API and can be used with Docker in Swarm Mode to monitor service availability.

## Configuration

You can use environment variables to configure Swarmonitor, except for container sets which use Yaml files.

### Environment variables

- SWMON_ORIGIN: String to identify server/cluster being monitored. It's sent along with the notification e-mail.
- SWMON_CONTAINER_DIR: Path to directory with container sets config files. Default: */etc/swarmonitor/containers*.
- SWMON_TICK_MINUTES: Interval in minutes to periodically check container sets. Default: 1.
- SWMON_SMTP_FROM: Sender e-mail address used to send notification e-mails.
- SWMON_SMTP_TO: E-mail address that will receive e-mail notifications.
- SWMON_SMTP_ADDRESS: STMP server address used to send e-mails.
- SWMON_SMTP_PORT: SMTP server port. Default: 587.
- SWMON_SMTP_USERNAME: User to authenticate to SMTP server.
- SWMON_SMTP_PASSWORD: Password to authenticate to SMTP server.
- SWMON_SMTP_DOMAIN: SMTP domain (hello).
- SWMON_SMTP_AUTH: SMTP authentication method. Default: *plain*.

### Container sets

Create a Yaml file in the container directory for each set of containers you want to monitor. Name each file with a string that will help you identify which containers are being checked.

Each Yaml file must contain simple matchings of string keys to string values. They should match labels/label values set in all containers that are part of that container set.

Instead of just checking if a container is running, you can check if any container in a container set is healthy according to Docker healthchecks. Just set the file extension to *.healthy.yml*.

## Example

A file */etc/swarmonitor/containers/webserver.healthy.yml* in your container config dir, containing the following data:

```
com.docker.swarm.service.name: site_webserver
environment: production
```

When Swarmonitor runs, it will check if there are any running containers matching the labels com.docker.swarm.service.name = site_webserver and environment = production. If there is at least one, it will consider this container set to be available and OK. If there is none, it will send an e-mail notification.

In this case, it will check if there is at least one running container that Docker considers healthy because the filename ends on ".healthy.yml".

The label "com.docker.swarm.service.name" can be used to monitor Docker Swarm services and make sure at least one container is running in that service. This label is automatically created by Docker Swarm when deploying services.
