import {
	request
} from "../libs/utils.js"

import {
	dayName,
	hijriDate,
	isoTimeString,
	isoDateString,
	isValidDate,
	fullDate,
} from "../libs/datetime.js"

function app() {
	let state = {
		time: new Date(),

		images: [],
		events: [],
		targets: [],

		nextEvent: -1,
		activeImage: -1,
		currentTarget: -1,

		beep: null,
		debugMode: false,
	}

	// API methods
	function loadData() {
		request("/api/data", "1m")
			.then(response => {
				// Save images
				state.images = response.images

				// Save event times
				let events = [], targets = []
				response.events.forEach(e => {
					let hasIqama = typeof (e.iqama) == "number" && e.iqama > 0,
						eventTime = new Date(e.time),
						iqamaTime = hasIqama ? new Date(e.iqama) : null

					if (e.name !== "nextFajr") {
						events.push({ name: e.name, time: eventTime, iqama: iqamaTime })
					}

					targets.push({ name: e.name, time: eventTime })
					if (hasIqama) {
						targets.push({ name: `${e.name}Iqama`, time: iqamaTime })
					}
				})

				state.events = events
				state.targets = targets

				// Create timeout to load data again tomorrow
				let time = new Date(),
					minutes = time.getHours() * 60 + time.getMinutes(),
					minutesTillNextDay = 24 * 60 - minutes + 1,
					msTillNextDay = minutesTillNextDay * 60 * 1000

				console.log(`Will reload data in ${minutesTillNextDay} minutes`)
				setTimeout(loadData, msTillNextDay)
			})
			.catch(err => { alert(err.message) })
			.finally(() => { m.redraw() })
	}

	// Local methods
	function startImages() {
		state.activeImage++
		if (state.activeImage >= state.images.length) {
			state.activeImage = 0
		}

		m.redraw()
		setTimeout(startImages, 20 * 1000)
	}

	function startClock() {
		// Calculate current time
		if (state.debugMode) {
			state.time.setSeconds(state.time.getSeconds() + 1)
		} else {
			state.time = new Date()
		}

		// Find next event
		state.nextEvent = state.events.findIndex(event => {
			let iqamaTime = event.iqama || event.time
			return event.time > state.time || iqamaTime > state.time
		})

		// Find next target
		let oldTarget = state.currentTarget
		state.currentTarget = state.targets.findIndex(target => {
			return target.time > state.time
		})

		// If target changed trigger the alarm
		if (oldTarget !== -1 && state.currentTarget !== oldTarget) {
			state.beep.play()
		}

		m.redraw()
		setTimeout(startClock, 1000)
	}

	function getCountdown() {
		if (state.currentTarget < 0 || state.currentTarget >= state.targets.length) {
			return ""
		}

		let target = state.targets[state.currentTarget],
			timeDiff = target.time - state.time,
			diffSeconds = timeDiff / 1000,
			diffMinutes = 0,
			diffHours = 0,
			text = ""

		if (diffSeconds >= 60 * 60) {
			diffHours = Math.floor(diffSeconds / 3600)
			diffMinutes = Math.floor((diffSeconds - diffHours * 3600) / 60)
			diffSeconds = Math.floor(diffSeconds - diffHours * 3600 - diffMinutes * 60)
		} else if (diffSeconds >= 60) {
			diffMinutes = Math.floor(diffSeconds / 60)
			diffSeconds = Math.floor(diffSeconds - diffMinutes * 60)
		}

		if (diffHours > 0) {
			if (diffMinutes === 0) text = `${diffHours} jam lagi`
			else text = `${diffHours} jam ${diffMinutes} menit lagi`
		} else if (diffMinutes > 0) {
			if (diffSeconds === 0) text = `${diffMinutes} menit lagi`
			else text = `${diffMinutes} menit ${diffSeconds} detik lagi`
		} else {
			text = `${diffSeconds} detik lagi`
		}

		return `${getTargetName(target.name)} ${text}`
	}

	function getEventName(name) {
		switch (name) {
			case "fajr": return "Subuh"
			case "sunrise": return "Syuruq"
			case "zuhr": return "Zuhur"
			case "asr": return "Ashar"
			case "maghrib": return "Maghrib"
			case "isha": return "Isha"
			case "nextFajr": return "Subuh"
		}
	}

	function getTargetName(name) {
		let prefix = ""
		if (name.endsWith("Iqama")) {
			prefix = "Iqamah "
			name = name.replace(/Iqama$/, "")
		}

		return prefix + getEventName(name)
	}

	// Lifecycle methods
	function renderView() {
		// Prepare variables
		let day = dayName(state.time.getDay()),
			strTime = isoTimeString(state.time, true),
			strDate = fullDate(state.time),
			strHijri = hijriDate(state.time),
			activeImage = state.images[state.activeImage],
			appAttributes = {},
			appContents = []

		// If there is active image, use it
		if (activeImage != null) {
			appContents = [
				m("img#main-image", {
					src: activeImage.url,
					loading: "lazy",
				}),
			]

			appAttributes = {
				style: {
					"--main-color": activeImage.mainColor,
					"--header-main": activeImage.headerMain,
					"--header-accent": activeImage.headerAccent,
					"--header-color": activeImage.headerFont,
					"--footer-main": activeImage.footerMain,
					"--footer-accent": activeImage.footerAccent,
					"--footer-color": activeImage.footerFont,
				},
				onclick() {
					state.debugMode = false
					state.time = new Date()
					return false
				}
			}
		}

		// Populate app contents
		let eventBoxes = state.events.map((e, idx) => {
			// Prepare box contents
			let boxContents = [
				m("p.event__name", getEventName(e.name)),
				m("p.event__time", isoTimeString(e.time)),
			]

			if (isValidDate(e.iqama)) boxContents.push(
				m("p.event__iqama", isoTimeString(e.iqama)),
			)

			// Prepare box attributes
			let boxAttributes = {
				class: idx === state.nextEvent ? "event--target" : null,
				onclick() {
					let eventTime = new Date(e.time.getTime() - 15 * 1000)
					state.debugMode = true
					state.time = eventTime
					return false
				}
			}

			return m(".event", boxAttributes, boxContents)
		})

		appContents.push(
			m("#header",
				m("p#clock", strTime),
				m("p#date", `${day}, ${strDate} M / ${strHijri} H`),
				m("p#countdown", getCountdown()),
			),
			m("#footer",
				m(".footer__space"),
				...eventBoxes,
				m(".footer__space"),
			),
		)

		return m("#app", appAttributes, appContents)
	}

	function onInit() {
		loadData()
		startClock()
		startImages()
		state.beep = new Audio("/res/beep.wav")
	}

	return {
		view: renderView,
		oninit: onInit,
	}
}

function startApp() {
	m.mount(document.body, app)
}

startApp()