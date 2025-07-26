FROM node:18-alpine

WORKDIR /app

# Copy package files first for better caching
COPY package*.json ./
RUN npm install

# Copy application code
COPY . .

# Expose port
EXPOSE 3000

# Start development server
CMD ["npm", "start"]
