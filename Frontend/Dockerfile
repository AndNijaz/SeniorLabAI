FROM node:20

WORKDIR /app

COPY package.json .
COPY package-lock.json .
RUN npm install
RUN npm --version

COPY . .

EXPOSE 3000

CMD ["npm", "run", "dev"]
