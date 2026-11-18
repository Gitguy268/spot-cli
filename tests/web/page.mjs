// Opens site/index.html from file:// in Chromium at desktop and phone sizes,
// clicks the buttons, and fails on console errors or sideways scrolling.
// PLAYWRIGHT=/path/to/node_modules/playwright/index.mjs node tests/web/page.mjs [shotdir]
// Set CHROMIUM to use a Chromium other than the one that Playwright downloaded.
import { pathToFileURL } from "node:url";
import { resolve } from "node:path";
const { chromium } = await import(process.env.PLAYWRIGHT || "playwright");
const url = pathToFileURL(resolve("site/index.html")).href;
const shots = process.argv[2];
const browser = await chromium.launch(process.env.CHROMIUM ? { executablePath: process.env.CHROMIUM } : {});
let failed = false;
for (const [name, vp, scheme] of [["desktop", { width: 1280, height: 900 }, "dark"], ["phone", { width: 390, height: 844 }, "light"]]) {
  const page = await browser.newPage({ viewport: vp, colorScheme: scheme });
  const errors = [];
  page.on("console", m => m.type() === "error" && errors.push(m.text()));
  page.on("pageerror", e => errors.push(String(e)));
  await page.goto(url);
  await page.waitForTimeout(1500);
  for (const act of ["pet", "fail", "pass", "feed", "trick"]) {
    await page.click(`[data-act="${act}"]`);
    await page.waitForTimeout(400);
  }
  const caption = await page.textContent("#caption");
  await page.click('[data-tab="fish"]');
  const fishShown = await page.isVisible('[data-name="fish"]');
  const sideways = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 1);
  const painted = await page.evaluate(() => {
    const c = document.getElementById("room"), d = c.getContext("2d").getImageData(0, 0, c.width, c.height).data;
    let n = 0; for (let i = 3; i < d.length; i += 4) if (d[i]) n++; return n;
  });
  await page.waitForTimeout(6000);
  const minis = await page.$$eval(".react canvas", cs => cs.length);
  if (shots) await page.screenshot({ path: `${shots}/page-${name}.png`, fullPage: name === "phone" });
  const ok = !errors.length && !sideways && fishShown && painted > 300 && minis >= 1 && caption.length > 0;
  console.log(`${ok ? "ok  " : "FAIL"} ${name}: errors=${errors.length} sideways=${sideways} pixels=${painted} reactions=${minis} caption="${caption}"`);
  errors.forEach(e => console.log("   ", e));
  failed ||= !ok;
  await page.close();
}
await browser.close();
process.exit(failed ? 1 : 0);
