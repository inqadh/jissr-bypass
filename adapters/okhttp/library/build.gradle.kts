plugins {
    id("com.android.library")
}

android {
    namespace = "org.jissr.bypass.okhttp"
    compileSdk = 36

    defaultConfig {
        minSdk = 26
        consumerProguardFiles("consumer-rules.pro")
    }

    testOptions {
        unitTests.isReturnDefaultValues = true
    }

    publishing {
        singleVariant("release") {
            withSourcesJar()
        }
    }
}

dependencies {
    val localAar = providers.gradleProperty("jissrBypassAar")
    if (localAar.isPresent) {
        compileOnly(files(localAar.get()))
        testImplementation(files(localAar.get()))
    } else {
        api(
            "org.jissr:jissr-bypass-android:" +
                providers.gradleProperty("jissrBypassVersion").get()
        )
    }
    api("com.squareup.okhttp3:okhttp:5.3.0")

    testImplementation("junit:junit:4.13.2")
    testImplementation("com.squareup.okhttp3:mockwebserver:5.3.0")
    testImplementation("org.json:json:20250517")
}
