fs = require('fs');

base = require('./base');
site = require('./site');

const main = async () => {
  // INTERNAL OPTION
  options = { 
    browserType: "firefox", // "chrome, firefox"
    headless: false,
    screenshot: true, 
    screenshotDir: "./screenshots"
  }
  const browser = await base.browser_init(options);
  options.version = await browser.version();
  console.log(options);

  await site.process_login(browser, options);
  data = {}
  await site.process_completed(browser, data);
  console.log(data);
  
  await site.process_logout(browser, options);
  await base.browser_close(browser);
};

main();

  
  