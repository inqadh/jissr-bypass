plugins {
    id("com.android.application")
}

android {
    namespace = "org.jissr.bypass.example"
    compileSdk = 36

    defaultConfig {
        applicationId = "org.jissr.bypass.example"
        minSdk = 26
        targetSdk = 36
        versionCode = 1
        versionName = "1.0"
    }
}

dependencies {
    implementation(
        "org.jissr:jissr-bypass-android:" +
            providers.gradleProperty("jissrBypassVersion").get()
    )
    implementation("androidx.activity:activity-ktx:1.11.0")
    implementation("androidx.lifecycle:lifecycle-runtime-ktx:2.9.1")
}
