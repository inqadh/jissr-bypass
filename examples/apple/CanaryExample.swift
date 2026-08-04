import Foundation
import Jissrbypass

func verifyJissrBypass() async throws -> String {
    try await Task.detached(priority: .userInitiated) {
        var error: NSError?
        let value = JissrbypassNewClient()
            .verifyCanary("https://blocked.yusef.ren/v1/canary", error: &error)
        if let error {
            throw error
        }
        return value
    }.value
}
