FROM eclipse-temurin:21-jdk AS build
COPY . /app
WORKDIR /app
RUN ./mvnw package -DskipTests

FROM eclipse-temurin:21-jre
COPY --from=build /app/target/url-shortener-0.0.1-SNAPSHOT.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]