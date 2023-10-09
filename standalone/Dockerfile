FROM ghcr.io/autobrr/autobrr:latest

# Set build args 
ARG imdlVersion=0.1.12

# Set working directory to /app
WORKDIR /app

# Grab intermodal release and put it in its own folder
RUN mkdir imdlDir
RUN wget -O imdlDir/imdl.tgz "https://github.com/casey/intermodal/releases/download/v${imdlVersion}/imdl-v${imdlVersion}-x86_64-unknown-linux-musl.tar.gz"

# Extract the binary, move it to working directory, and delete everything else
RUN tar xzvf imdlDir/imdl.tgz -C imdlDir
RUN mv imdlDir/imdl .
RUN rm -rf imdlDir

# Copy script to working directory
COPY check-rar.sh .
