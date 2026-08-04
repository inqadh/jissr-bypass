package org.jissr.bypass.example

import android.os.Bundle
import android.widget.TextView
import androidx.activity.ComponentActivity
import androidx.lifecycle.lifecycleScope
import jissrbypass.Jissrbypass
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        val result = TextView(this).also { setContentView(it) }
        lifecycleScope.launch {
            result.text = withContext(Dispatchers.IO) {
                try {
                    Jissrbypass.newClient()
                        .verifyCanary("https://blocked.yusef.ren/v1/canary")
                } catch (error: Exception) {
                    "Canary failed: ${error.message}"
                }
            }
        }
    }
}
