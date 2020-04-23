fs = require('fs');

base = require('./base');
site = require('./site');

const main = async () => {
  // INTERNAL OPTION
  options = { 
    browserType: "firefox", // "chrome, firefox"
    headless: false,
    screenshot: true, 
    screenshotDir: "./screenshots",
    lighthouse: false, 
    lighthouseDir: "./screenshots"
  }
  const browser = await base.browser_init(options);
  options.version = await browser.version();
  console.log(options);
  if (options.lighthouse) {
    console.log("You can use the following site to view lighthouse reports:");
    console.log("https://googlechrome.github.io/lighthouse/viewer/");
  }

  await site.process_login(browser, options);
  await site.process_completed(browser, options);

  // // EAMPLE: cycle thru each tab 3 times
  // betweenTabsMs = 2000;
  // count = 1;
  // for (i=0; i<options.loops; i++) {
  //   await site.process_control_tab(browser, options);
  //   await base.delay(betweenTabsMs);

  //   await site.process_servers_tab(browser, options);
  //   await base.delay(betweenTabsMs);

  //   await site.process_files_tab(browser, options);
  //   await base.delay(betweenTabsMs);

  //   await site.process_logs_tab(browser, options);
  //   await base.delay(betweenTabsMs);

  //   await site.process_reports_tab(browser, options);
  //   await base.delay(betweenTabsMs);

  //   await site.process_setup_tab(browser, options);  
  //   await base.delay(betweenTabsMs);

  //   await site.process_monitor_tab(browser, options);  
  //   await base.delay(betweenTabsMs);
  // }
  
  await site.process_logout(browser, options);
  await base.browser_close(browser);
};

main();

  
  