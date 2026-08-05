// swift-tools-version: 5.10
import PackageDescription

let package = Package(
    name: "JissrBypassAppleAdapter",
    platforms: [
        .iOS(.v15),
        .macOS(.v12),
    ],
    products: [
        .library(
            name: "JissrBypassURLSession",
            targets: ["JissrBypassURLSession"]
        ),
    ],
    targets: [
        .binaryTarget(
            name: "JissrBypass",
            path: "Binaries/JissrBypass.xcframework"
        ),
        .target(
            name: "JissrBypassURLSession",
            dependencies: ["JissrBypass"]
        ),
        .testTarget(
            name: "JissrBypassURLSessionTests",
            dependencies: ["JissrBypassURLSession"]
        ),
    ]
)
