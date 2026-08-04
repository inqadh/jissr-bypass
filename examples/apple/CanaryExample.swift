import Foundation
import Jissrbypass

struct BypassResponse {
    let statusCode: Int
    let body: Data
    let headersJSON: String
}

func fetchWithJissrBypass() async throws -> BypassResponse {
    try await Task.detached(priority: .userInitiated) {
        var error: NSError?
        let response = JissrbypassNewClient().execute(
            "GET",
            url: "https://blocked.yusef.ren/v1/health",
            body: Data(),
            headersJSON: #"{"Accept":"application/json"}"#,
            error: &error
        )
        if let error {
            throw error
        }
        guard let response else {
            throw NSError(domain: "JissrBypass", code: -1)
        }
        return BypassResponse(
            statusCode: Int(response.getStatusCode()),
            body: response.getBody(),
            headersJSON: response.getHeadersJSON()
        )
    }.value
}
