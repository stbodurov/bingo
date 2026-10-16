# 1) React client
FROM node:22-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY web/ ./
RUN npx tsc -b && npx vite build --outDir /static

# 2) Spring Boot jar
FROM gradle:8.14-jdk21 AS app
WORKDIR /src
COPY settings.gradle.kts build.gradle.kts ./
COPY src ./src
COPY --from=web /static ./src/main/resources/static
RUN SKIP_WEB=1 gradle bootJar --no-daemon -q

# 3) Runtime
FROM eclipse-temurin:21-jre
RUN useradd -r -u 10001 sundial && mkdir -p /data && chown sundial /data
WORKDIR /data
COPY --from=app /src/build/libs/sundial-1.0.0.jar /app/sundial.jar
USER sundial
ENV SUNDIAL_DB_URL="jdbc:h2:file:/data/sundial;MODE=PostgreSQL" PORT=8080
EXPOSE 8080
ENTRYPOINT ["java", "-XX:MaxRAMPercentage=75", "-jar", "/app/sundial.jar"]
