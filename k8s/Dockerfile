# Start from the Ubuntu base image
FROM ubuntu:latest

# Set the timezone to UTC
ENV TZ=UTC

# Install necessary packages
RUN apt-get update && apt-get install -y \
    cron \
    smartmontools \
    && apt-get clean

# Add your shell script to the image
COPY disktool.sh /usr/local/bin/disktool.sh
RUN chmod +x /usr/local/bin/disktool.sh

# Run the cron service and tail the cron log
# CMD ["bash", "-c", "touch /var/log/cron.log && cron && tail -f /var/log/cron.log"]








# ---------------------------------------
#     # Start from the Ubuntu base image
# FROM ubuntu:latest

# # Install necessary packages
# RUN apt-get update && apt-get install -y \
#     cron \
#     && apt-get clean

# # Add your shell script to the image
# COPY disktool.sh /usr/local/bin/disktool.sh
# RUN chmod +x /usr/local/bin/disktool.sh

# # Set up the environment variable for the cron schedule
# # Default is every 2 minutes (*/2 * * * *)
# ENV CRON_SCHEDULE="*/2 * * * *"

# # Create the script that updates the cron job based on the environment variable
# RUN echo "#!/bin/bash\necho \"$CRON_SCHEDULE /bin/bash /usr/local/bin/disktool.sh >> /var/log/cron.log 2>&1\" > /etc/cron.d/my_cron && \
#     chmod 0644 /etc/cron.d/my_cron && \
#     crontab /etc/cron.d/my_cron" > /usr/local/bin/setup_cron.sh && \
#     chmod +x /usr/local/bin/setup_cron.sh

# # Run the cron service and ensure the setup script runs at container startup
# CMD ["/bin/bash", "-c", "/usr/local/bin/setup_cron.sh && cron && tail -f /var/log/cron.log"]
