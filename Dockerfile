# Use the official Golang image as the base image
FROM ruby

# Copy everything to /app, it's better to use a mount though
COPY . /app

# Install bash for interactive shell access
RUN apt-get update && apt-get install -y bash golang vim iputils-ping

# RVM
RUN curl -sSL https://get.rvm.io | bash
RUN echo "source /etc/profile.d/rvm.sh" >> ~/.bashrc

# Packages for Rails apps
RUN apt-get install -y libldap2-dev libidn11-dev

# Set the working directory
WORKDIR /app

# Start a shell when the container starts
CMD ["/bin/bash"]