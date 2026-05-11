# Build this from the examples/gcp-cloudrun directory!
# Command: docker build -t gcp-cloudrun-example .

FROM node:20-alpine

WORKDIR /app

# Copy package files and install dependencies
COPY package.json ./
RUN npm install

# Copy the application source code
COPY server.js ./
COPY public ./public

# Set standard Node.js environment variables
ENV NODE_ENV=production
ENV PORT=8080

EXPOSE 8080

CMD ["npm", "start"]
