import Foundation
import XCTest
@testable import JissrBypassURLSession

final class JissrBypassURLSessionTests: XCTestCase {
    func testMatchesOnlyExactConfiguredHost() throws {
        JissrBypassURLProtocol.configure(
            hosts: ["api.example.com"],
            executor: { _ in
                AdapterResponse(
                    statusCode: 200,
                    headersJSON: "{}",
                    body: Data()
                )
            }
        )

        let exact = URLRequest(url: URL(string: "https://api.example.com/v1")!)
        let subdomain = URLRequest(url: URL(string: "https://www.api.example.com/v1")!)
        let plaintext = URLRequest(url: URL(string: "http://api.example.com/v1")!)

        XCTAssertTrue(JissrBypassURLProtocol.canInit(with: exact))
        XCTAssertFalse(JissrBypassURLProtocol.canInit(with: subdomain))
        XCTAssertFalse(JissrBypassURLProtocol.canInit(with: plaintext))
    }

    func testRejectsNonstandardHTTPSPort() throws {
        JissrBypassURLProtocol.configure(
            hosts: ["api.example.com"],
            executor: { _ in
                AdapterResponse(
                    statusCode: 200,
                    headersJSON: "{}",
                    body: Data()
                )
            }
        )
        let request = URLRequest(url: URL(string: "https://api.example.com:8443/v1")!)
        XCTAssertFalse(JissrBypassURLProtocol.canInit(with: request))
    }
}
