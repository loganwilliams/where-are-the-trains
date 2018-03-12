FROM node:9.4.0-alpine

WORKDIR /client

COPY yarn.lock ./

RUN yarn

COPY . .

EXPOSE 3000

CMD ["yarn", "start"]