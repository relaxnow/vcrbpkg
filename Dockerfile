# Use the official Golang image as the base image
FROM ruby

# Copy everything to /app, it's better to use a mount though
COPY . /app

# See also: https://www.garron.me/en/linux/install-ruby-2-3-3-ubuntu.html
RUN echo "deb http://deb.debian.org/debian-security/ bookworm-security main" >> /etc/apt/sources.list

# Install useful tooling
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