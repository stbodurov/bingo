plugins {
    java
    id("org.springframework.boot") version "3.5.6"
    id("io.spring.dependency-management") version "1.1.7"
}

group = "com.sundial"
version = "1.0.0"

java {
    toolchain { languageVersion = JavaLanguageVersion.of(21) }
}

repositories { mavenCentral() }

dependencies {
    implementation("org.springframework.boot:spring-boot-starter-web")
    implementation("org.springframework.boot:spring-boot-starter-validation")
    implementation("org.springframework.boot:spring-boot-starter-data-jpa")
    implementation("org.springframework.boot:spring-boot-starter-security")
    implementation("org.springframework.boot:spring-boot-starter-actuator")
    implementation("org.flywaydb:flyway-core")
    runtimeOnly("com.h2database:h2")
    runtimeOnly("org.postgresql:postgresql")
    runtimeOnly("org.flywaydb:flyway-database-postgresql")
    testImplementation("org.springframework.boot:spring-boot-starter-test")
    testImplementation("org.springframework.security:spring-security-test")
    testRuntimeOnly("org.junit.platform:junit-platform-launcher")
}

tasks.withType<Test> { useJUnitPlatform() }

// Build the React client into the jar's static resources when Node is available.
val webDir = layout.projectDirectory.dir("web")
val npmInstall by tasks.registering(Exec::class) {
    workingDir = webDir.asFile
    commandLine("npm", "ci", "--no-audit", "--no-fund")
    inputs.file(webDir.file("package-lock.json"))
    outputs.dir(webDir.dir("node_modules"))
    onlyIf { System.getenv("SKIP_WEB") == null }
}
val webBuild by tasks.registering(Exec::class) {
    dependsOn(npmInstall)
    workingDir = webDir.asFile
    commandLine("npm", "run", "build")
    inputs.dir(webDir.dir("src"))
    inputs.file(webDir.file("index.html"))
    outputs.dir(layout.projectDirectory.dir("src/main/resources/static"))
    onlyIf { System.getenv("SKIP_WEB") == null }
}
tasks.named("processResources") { dependsOn(webBuild) }
