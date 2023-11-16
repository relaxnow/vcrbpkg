# Use the official Golang image as the base image
FROM ruby

# Copy everything to /app, it's better to use a mount though
COPY . /app

# Install bash for interactive shell access
RUN apt-get update && apt-get install -y bash golang

# Set the working directory
WORKDIR /app

# Start a shell when the container starts
CMD ["/bin/bash"]