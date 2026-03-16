/**
 * Standalone smoke script — verifies React #310 hooks-violation fix.
 * Run directly: npx tsx e2e/mission-hooks-smoke.ts
 */
import { chromium } from '@playwright/test'

async function main() {
  const browser = await chromium.launch({
    args: [
      '--host-resolver-rules=MAP scouter.dev.local 127.0.0.1',
      '--ignore-certificate-errors',
    ],
    headless: true,
  })

  const page = await browser.newPage()

  const errors: string[] = []
  page.on('console', (msg) => {
    if (msg.type() === 'error') {
      errors.push(msg.text())
    }
  })
  page.on('pageerror', (err) => {
    errors.push(`[pageerror] ${err.message}`)
  })

  console.log('Navigating to mission page...')
  await page.goto('https://scouter.dev.local/missions/summer-holiday-2026', {
    waitUntil: 'domcontentloaded',
    timeout: 20000,
  })

  // Dismiss onboarding overlay if present
  const skipBtn = page.locator('button:has-text("SKIP"), button:has-text("Skip")')
  if (await skipBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
    await skipBtn.click()
  }

  // Wait for React to render content
  await page.waitForTimeout(4000)

  const react310 = errors.filter(
    (e) =>
      e.includes('310') ||
      e.includes('Cannot update a component') ||
      e.includes('rendered more hooks') ||
      e.includes('Rendered fewer hooks') ||
      e.includes('Cannot read properties of null') ||
      e.includes('ErrorBoundary'),
  )

  console.log('\n=== Console Errors ===')
  if (errors.length === 0) {
    console.log('No console errors.')
  } else {
    errors.forEach((e) => console.log(' -', e))
  }

  console.log('\n=== React / Crash Errors ===')
  if (react310.length === 0) {
    console.log('✅ PASS — No React crashes or hooks violations found.')
  } else {
    console.log('❌ FAIL — React errors detected:')
    react310.forEach((e) => console.log(' -', e))
    process.exitCode = 1
  }

  // Take screenshot
  await page.screenshot({ path: 'e2e/screenshots/mission-hooks-smoke.png', fullPage: false })
  console.log('\nScreenshot saved to e2e/screenshots/mission-hooks-smoke.png')

  await browser.close()
}

main().catch((err) => {
  console.error(err)
  process.exit(1)
})
