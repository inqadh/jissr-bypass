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
                    val response = Jissrbypass.newClient().execute(
                        "GET",
                        "https://blocked.yusef.ren/v1/health",
                        ByteArray(0),
                        """{"Accept":"application/json"}"""
                    )
                    "HTTP ${response.statusCode}: " +
                        response.body.toString(Charsets.UTF_8)
                } catch (error: Exception) {
                    "Request failed: ${error.message}"
                }
            }
        }
    }
}
