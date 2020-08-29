(() => {
  // internal/view/js/libs/utils.js
  function timeout(duration, promise) {
    let ms = duration;
    if (typeof duration == "string") {
      let unitName = duration.replace(/^\d+/, "");
      ms = parseInt(duration, 10);
      switch (unitName) {
        case "s":
          ms *= 1e3;
          break;
        case "M":
          ms *= 60 * 1e3;
          break;
        case "H":
          ms *= 60 * 60 * 1e3;
          break;
        default:
          ms = 0;
      }
    }
    if (ms === 0)
      return promise;
    else
      return new Promise((resolve, reject) => {
        setTimeout(() => reject(new Error(`Timeout after ${duration}`)), ms);
        promise.then(resolve, reject);
      });
  }
  async function request(url, ms, options) {
    let fetchRequest = fetch(url, options), response = await timeout(ms, fetchRequest);
    if (!response.ok) {
      let responseText = await response.text();
      throw Error(`${responseText.trim()} (${response.status})`);
    }
    if (response.headers.get("content-type") === "application/json") {
      return await response.json();
    }
    return await response.text();
  }

  // internal/view/js/libs/datetime.js
  function isValidDate(d) {
    return d instanceof Date && !isNaN(d);
  }
  function isoTimeString(d, showSeconds) {
    if (!isValidDate(d))
      return "";
    let H = d.getHours(), m2 = d.getMinutes(), s = d.getSeconds(), strH = String(H).padStart(2, "0"), strM = String(m2).padStart(2, "0"), strS = String(s).padStart(2, "0");
    if (typeof showSeconds == "boolean" && showSeconds) {
      return `${strH}:${strM}:${strS}`;
    } else {
      m2 = s >= 30 ? m2 + 1 : m2;
      strM = String(m2).padStart(2, "0");
      return `${strH}:${strM}`;
    }
  }
  function isoDateString(d) {
    if (!isValidDate(d))
      return "";
    let date = String(d.getDate()).padStart(2, "0"), month = String(d.getMonth() + 1).padStart(2, "0"), year = d.getFullYear();
    return `${year}-${month}-${date}`;
  }
  function dayName(day) {
    switch (day) {
      case 0:
        return "Minggu";
      case 1:
        return "Senin";
      case 2:
        return "Selasa";
      case 3:
        return "Rabu";
      case 4:
        return "Kamis";
      case 5:
        return "Jum'at";
      case 6:
        return "Sabtu";
    }
  }
  function hijriDate(d) {
    if (!isValidDate(d))
      return "";
    return new Intl.DateTimeFormat("id-u-ca-islamic", {
      day: "2-digit",
      month: "2-digit",
      year: "2-digit"
    }).format(d).replace(/\s+H$/i, "");
  }

  // internal/view/js/entries/index.js
  function app() {
    let state = {
      time: new Date(),
      images: [],
      events: [],
      targets: [],
      nextEvent: -1,
      activeImage: -1,
      currentTarget: -1,
      beep: null
    };
    function loadData() {
      request("/api/data", "1m").then((response) => {
        state.images = response.images;
        let events = [], targets = [];
        response.events.forEach((e) => {
          let hasIqama = typeof e.iqama == "number" && e.iqama > 0, eventTime = new Date(e.time), iqamaTime = hasIqama ? new Date(e.iqama) : null;
          if (e.name !== "nextFajr") {
            events.push({name: e.name, time: eventTime, iqama: iqamaTime});
          }
          targets.push({name: e.name, time: eventTime});
          if (hasIqama) {
            targets.push({name: `${e.name}Iqama`, time: iqamaTime});
          }
        });
        state.events = events;
        state.targets = targets;
        let time = new Date(), minutes = time.getHours() * 60 + time.getMinutes(), minutesTillNextDay = 24 * 60 - minutes + 1, msTillNextDay = minutesTillNextDay * 60 * 1e3;
        console.log(`Will reload data in ${minutesTillNextDay} minutes`);
        setTimeout(loadData, msTillNextDay);
      }).catch((err) => {
        alert(err.message);
      }).finally(() => {
        m.redraw();
      });
    }
    function startImages() {
      state.activeImage++;
      if (state.activeImage >= state.images.length) {
        state.activeImage = 0;
      }
      m.redraw();
      setTimeout(startImages, 20 * 1e3);
    }
    function startClock() {
      let currentTime = new Date(), nextEvent = state.events.findIndex((event) => {
        let iqamaTime = event.iqama || event.time;
        return event.time > currentTime || iqamaTime > currentTime;
      }), newTarget = state.targets.findIndex((target) => {
        return target.time > currentTime;
      });
      if (newTarget !== state.currentTarget) {
        state.beep.play();
      }
      state.time = currentTime;
      state.nextEvent = nextEvent;
      state.currentTarget = newTarget;
      m.redraw();
      setTimeout(startClock, 1e3);
    }
    function getCountdown() {
      if (state.currentTarget < 0 || state.currentTarget >= state.targets.length) {
        return "";
      }
      let target = state.targets[state.currentTarget], timeDiff = target.time - state.time, diffSeconds = timeDiff / 1e3, diffMinutes = 0, diffHours = 0, text = "";
      if (diffSeconds >= 60 * 60) {
        diffHours = Math.floor(diffSeconds / 3600);
        diffMinutes = Math.floor((diffSeconds - diffHours * 3600) / 60);
        diffSeconds = diffSeconds - diffHours * 3600 - diffMinutes * 60;
      } else if (diffSeconds >= 60) {
        diffMinutes = Math.floor(diffSeconds / 60);
        diffSeconds = diffSeconds - diffMinutes * 60;
      }
      if (diffHours > 0) {
        if (diffMinutes === 0)
          text = `${diffHours} jam lagi`;
        else
          text = `${diffHours} jam ${diffMinutes} menit lagi`;
      } else if (diffMinutes > 0) {
        if (diffSeconds === 0)
          text = `${diffMinutes} menit lagi`;
        else
          text = `${diffMinutes} menit ${diffSeconds} detik lagi`;
      } else {
        text = `${diffSeconds} detik lagi`;
      }
      return `${getTargetName(target.name)} ${text}`;
    }
    function getEventName(name) {
      switch (name) {
        case "fajr":
          return "Subuh";
        case "sunrise":
          return "Syuruq";
        case "zuhr":
          return "Zuhur";
        case "asr":
          return "Ashar";
        case "maghrib":
          return "Maghrib";
        case "isha":
          return "Isha";
      }
    }
    function getTargetName(name) {
      let prefix = "";
      if (name.endsWith("Iqama")) {
        prefix = "Iqamah ";
        name = name.replace(/Iqama$/, "");
      }
      return prefix + getEventName(name);
    }
    function renderView() {
      let day = dayName(state.time.getDay()), strTime = isoTimeString(state.time, true), strDate = isoDateString(state.time).replace(/\//g, "-"), strHijri = hijriDate(state.time).replace(/\//g, "-"), activeImage = state.images[state.activeImage], appAttributes = {}, appContents = [];
      if (activeImage != null) {
        appContents = [
          m("img#main-image", {
            src: activeImage.url,
            loading: "lazy"
          })
        ];
        appAttributes = {
          style: {
            "--main-color": activeImage.mainColor,
            "--header-main": activeImage.headerMain,
            "--header-accent": activeImage.headerAccent,
            "--header-color": activeImage.headerFont,
            "--footer-main": activeImage.footerMain,
            "--footer-accent": activeImage.footerAccent,
            "--footer-color": activeImage.footerFont
          }
        };
      }
      let eventBoxes = state.events.map((e, idx) => {
        let className = idx === state.nextEvent ? "event--target" : null, eventContents = [
          m("p.event__name", getEventName(e.name)),
          m("p.event__time", isoTimeString(e.time))
        ];
        if (isValidDate(e.iqama))
          eventContents.push(m("p.event__iqama", isoTimeString(e.iqama)));
        return m(".event", {class: className}, eventContents);
      });
      appContents.push(m("#header", m("p#clock", strTime), m("p#date", `${day}, ${strDate} M / ${strHijri} H`), m("p#countdown", getCountdown())), m("#footer", m(".footer__space"), ...eventBoxes, m(".footer__space")));
      appAttributes.onclick = () => {
        state.beep.play();
      };
      return m("#app", appAttributes, appContents);
    }
    function onInit() {
      loadData();
      startClock();
      startImages();
      state.beep = new Audio("/res/beep.wav");
    }
    return {
      view: renderView,
      oninit: onInit
    };
  }
  function startApp() {
    m.mount(document.body, app);
  }
  startApp();
})();
