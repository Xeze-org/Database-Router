# Build this from the root of the repository!
# Command: docker build -f examples/gcp-cloudrun/Dockerfile -t gcp-cloudrun-example .

FROM node:20-alpine

WORKDIR /app

# Copy the local SDK dependencies first to allow npm install
# We need the dbr and core SDKs because they are referenced locally via "file:..."
COPY sdk/node ./sdk/node
COPY core/node ./core/node

# Copy the example app package.json
COPY examples/gcp-cloudrun/package.json ./examples/gcp-cloudrun/

# Install dependencies for the example app
WORKDIR /app/examples/gcp-cloudrun
RUN npm install

# Copy the application source code
COPY examples/gcp-cloudrun/server.js ./

# Set standard Node.js environment variables
ENV NODE_ENV=production
ENV PORT=8080

EXPOSE 8080

CMD ["npm", "start"]
