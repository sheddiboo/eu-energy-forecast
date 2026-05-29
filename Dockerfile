# 1. Use a lightweight Python base image
FROM python:3.11-slim

# 2. Prevent Python from buffering stdout so logs stream directly to AWS CloudWatch
ENV PYTHONUNBUFFERED=1
ENV PYTHONDONTWRITEBYTECODE=1

# 3. Install Golang and required system dependencies
RUN apt-get update && apt-get install -y \
    golang-go \
    curl \
    git \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# 4. Install uv (Python package manager)
RUN curl -LsSf https://astral.sh/uv/install.sh | sh
ENV PATH="/root/.cargo/bin:${PATH}"

# 5. Install Bruin CLI natively into the system binaries
RUN curl -sfL https://raw.githubusercontent.com/bruin-data/bruin/main/install.sh | sh -s -- -b /usr/local/bin

# 6. Set the working directory
WORKDIR /app

# 7. Copy your entire project into the container
COPY . /app

# 8. Define the single execution command
CMD ["bruin", "run", "."]