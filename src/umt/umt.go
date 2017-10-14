package umt

import (
	"accessdb"
	"fmt"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine
	"html"
	"io"
	"net/http"
	"os"
	"static"
	"strconv"
	"strings"
	"time"
)

func getDoctype() string {
	return `<!DOCTYPE html>
<html>
<head>
<meta charset=utf-8 />
`
}

func outputStaticFile(w http.ResponseWriter, filename string) {
	fullpath := "/home/ec2-user/wayneserver/staticappcontent/umt/" + filename
	fh, err := os.Open(fullpath)
	if err != nil {
		fmt.Println(err)
		panic("error opening static file, point 14")
	}
	_, err = io.Copy(w, fh)
	fh.Close()
	return
}

func mainPage(w http.ResponseWriter, r *http.Request, op string, userid uint64) {
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	// outputStaticFile(w, "umt.html")
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	getform := r.Form
	var paramsetid uint64
	var name string
	paramsetid = 0
	_, ok := getform["paramset"]
	if ok {
		paramsetstr := getform["paramset"][0]
		paramsetid, err = strconv.ParseUint(paramsetstr, 10, 64)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		// get the name
		sql := "SELECT name FROM umt_paramset WHERE (id_paramset = ?);"
		db := accessdb.GetDbConnection()
		defer db.Close()
		sel, err := db.Prepare(sql)
		if err != nil {
			fmt.Println(w)
			panic("Prepare failed")
		}
		sel.Bind(paramsetid)
		rows, _, err := sel.Exec()
		if err != nil {
			fmt.Println(w)
			panic("Exec failed")
		}
		for _, row := range rows {
			name = row.Str(0)
		}
	}
	// echo
	fmt.Fprintln(w, `
<!DOCTYPE html>
<html>
<head>
<meta charset=utf-8 />
<title>Ultimate Music Toy</title>
<link rel="stylesheet" href="jquery-ui.css" />
<script src="jquery-1.10.2.js"></script>
<script src="jquery-ui.js"></script>
<link rel="stylesheet" href="style.css" />
  <style>
  #tabdia label, #tabdia input { display:block; }
  #tabdia label { margin-top: 0.5em; }
  #tabdia input, #tabdia textarea { width: 95%; }
  #tabs { margin-top: 1em; }
  #tabs li .ui-icon-close { float: left; margin: 0.4em 0.2em 0 0; cursor: pointer; }
  #add_tab { cursor: pointer; }
  </style>

<script src="mt19937class.js" ></script>
<script>

// (C) 2011-2016 Wayne Radinsky

/*jslint browser: true, devel: true, passfail: true */
/*global MersenneTwister19937 */
/*global jQuery */
/*global Wad */
/*global WebSocket */

var globalAudioContext;
var gUmt;

// notes about terminology:
// lpnum -- loop number
// vcnum -- voice number
// "cursor" -- various cursors point to the end of an array + 1
//             (point where new notes will be added)
// instParams -- instrument parameters -- not used yet but will hold
//                instrument-specific parameters
// fpt -- "floating-point" -- floating point value of pitches for scale notes
//        (notes are expressed as both a ratio, num/dnom, and a floating-point
//        value where 2^fpt is proportional to the pitch
//        pitch values in gUmt.loop[lpnum].score.songTab[tab].voice[voicenum].notes[notenum].pitch
//        are RAW pitches and in floating-point format.

//    jsabort, cx, and ctstr are debugging functions that will be
//    REMOVED from the final product (assuming there is ever any such thing as a
//    final product :P )
//
//    Anything that starts with umt_test is test code and will be removed
//    (or at least shoved aside into some other file).
//
//    Actual music code starts with the function umtGetAudioContext


function jsabort(msg) {
    "use strict";
    window.alert(msg);
    throw new Error('Aborting javascript: ' + msg);
}

function cx(stringparameter) {
    "use strict";
    // this function exists just so I don't have to type 'console.log' over
    // plus prevents output from changing!
    console.log(stringparameter);
}

function ctstr(anyparameter) {
    "use strict";
    // a supposedly simple function to make a string from a var for debug
    //dumps, that has gotten way complicated over time...
    var x, rv, z, name, isArr;
    if (typeof anyparameter === 'object') {
        if (anyparameter === null) {
            return 'null';
        }
        name = '';
        isArr = false;
        if (anyparameter.constructor !== undefined) {
            name = anyparameter.constructor.toString();
        }
        if (name === 'function Object() { [native code] }') {
            name = '';
        }
        if (name === 'function Array() { [native code] }') {
            name = '';
            isArr = true;
        }
        if (name === 'function AudioContext() { [native code] }') {
            return 'object AudioContext';
        }
        if (name === '') {
            rv = '';
            for (z in anyparameter) {
                if (anyparameter.hasOwnProperty(z)) {
                    rv = rv + ', ' + z + ': ' + ctstr(anyparameter[z]);
                }
            }
            if (isArr) {
                return 'array [ ' + rv.substring(2) + ' ]';
            }
            return 'object { ' + rv.substring(2) + ' }';
        }
        return '{[Code Object]}';
    }
    if (typeof anyparameter === 'number') {
        return 'number ' + anyparameter.toString();
    }
    if (typeof anyparameter === 'string') {
        return 'string "' + anyparameter + '"';
    }
    if (typeof anyparameter === 'boolean') {
        return 'boolean ' + anyparameter.toString();
    }
    x = typeof anyparameter;
    window.alert(x);
    console.log('x', x);
    jsabort('typeof returned an undefined value');
}

// Actual music code starts here

function umtGetAudioContext() {
    "use strict";
    if (!window.AudioContext) {
        if (!window.webkitAudioContext) {
            window.alert("Your browser does not support the WebAudio API (HTML5).\nPlease use a compatible browser e.g. Google Chrome. If your WebAudio is disabled in Chrome, enable it in page about:flags");
            return;
        }
        window.AudioContext = window.webkitAudioContext;
    }
    var ctx = new window.AudioContext();
    return ctx;
}

// WAD has been modified to detect this global variable, and if it exists, use it instead of instantiating its own audio context
globalAudioContext = umtGetAudioContext();

</script>
<script src="wad.js"></script>
<script>

function umtSendLocalMsg(msg) {
    "use strict";
    if (gUmt.localSocket.readyState === 1) {
        gUmt.localSocket.send(msg);
    } else {
        console.log("Socket to local server is not connected.");
    }
}

function umtGetRando(seedNum) {
    "use strict";
    var mt1, initAry, x;
    mt1 = new MersenneTwister19937();
    initAry = [0x123, 0x234, 0x345, 0x456];
    mt1.initByArray(initAry, 4);
    for (x = 0; x < (seedNum * 32); x = x + 1) {
        mt1.genrandInt32();
    }
    return mt1;
}

// For all these "FM" building block functions, "revs" actually stands for "revolutions" -- i.e the number of times through the cycle
// The integer part is the number of complete revolutions so far and the fractional part is how far into the current revolution we are

function umtFMSine(revs, clipping, lean) {
    "use strict";
    var val, skip;
    if (clipping === 0) {
        if (lean !== 1.0) {
            skip = Math.floor(revs);
            revs = revs - skip;
            revs = Math.pow(revs, lean);
        }
        return Math.sin(gUmt.TAU * revs);
    }
    if (lean === 1.0) {
        val = Math.sin(gUmt.TAU * revs) * Math.exp(clipping);
    } else {
        skip = Math.floor(revs);
        revs = revs - skip;
        revs = Math.pow(revs, lean);
        val = Math.sin(gUmt.TAU * revs) * Math.exp(clipping);
    }
    if (val > 1.0) {
        return 1.0;
    }
    if (val < -1.0) {
        return -1.0;
    }
    return val;
}

function umtFMSquare(revs) {
    "use strict";
    var skip;
    skip = Math.floor(revs);
    revs = revs - skip;
    if (revs > 0.5) {
        return 1.0;
    }
    return -1.0;
}

function umtFMVarSquare(revs, variance) {
    "use strict";
    var skip;
    // variance == range 0.5 to 1.0
    skip = Math.floor(revs);
    revs = revs - skip;
    if (revs > variance) {
        return 1.0;
    }
    return -1.0;
}

function umtFMTriangle(revs) {
    "use strict";
    var skip;
    skip = Math.floor(revs);
    revs = revs - skip;
    if (revs > 0.5) {
        return 1.0 - ((revs - 0.5) * 4.0);
    }
    return (revs * 4.0) - 1.0;
}

function umtFMVarTriangle(revs, variance) {
    "use strict";
    var skip;
    // for variance, 0.5 == center == triangle
    skip = Math.floor(revs);
    revs = revs - skip;
    if (revs > variance) {
        return 1.0 - (((revs - variance) / (1.0 - variance)) * 2.0);
    }
    return ((revs / variance) * 2.0) - 1.0;
}

function umtFMSawtooth(revs) {
    "use strict";
    var skip;
    skip = Math.floor(revs);
    revs = revs - skip;
    return (revs - 0.5) * 2.0;
}

function umtGenerateANoteSineWave(frequency, duration, amplitude, clipping, lean) {
    "use strict";
    // consult https://dvcs.w3.org/hg/audio/raw-file/tip/webaudio/specification.html
    var numSamples, theBuffer, bufData, i, fade, sampleRate;
    sampleRate = gUmt.globalCtx.sampleRate;
    numSamples = duration * sampleRate;
    theBuffer = gUmt.globalCtx.createBuffer(1, numSamples, sampleRate); // numberOfChannels, length, sampleRate
    bufData = theBuffer.getChannelData(0);
    for (i = 0; i < numSamples; i = i + 1) {
        fade = 1.0 - (i / numSamples);
        bufData[i] = umtFMSine(i * (frequency / sampleRate), clipping, 1.0 + lean) * (amplitude * fade);
    }
    return theBuffer;
}

function umtGenerateANoteSineStartStopWave(frequency, duration, amplitude, startclipping, startlean, stopclipping, stoplean) {
    "use strict";
    // consult https://dvcs.w3.org/hg/audio/raw-file/tip/webaudio/specification.html
    var numSamples, theBuffer, bufData, i, fade, sampleRate, pctThere, clipGap, leanGap, currentClip, currentLean;
    sampleRate = gUmt.globalCtx.sampleRate;
    numSamples = duration * sampleRate;
    clipGap = stopclipping - startclipping;
    leanGap = stoplean - startlean;
    theBuffer = gUmt.globalCtx.createBuffer(1, numSamples, sampleRate); // numberOfChannels, length, sampleRate
    bufData = theBuffer.getChannelData(0);
    for (i = 0; i < numSamples; i = i + 1) {
        pctThere = i / numSamples;
        fade = 1.0 - pctThere;
        currentClip = (clipGap * pctThere) + startclipping;
        currentLean = (leanGap * pctThere) + startlean;
        bufData[i] = umtFMSine(i * (frequency / sampleRate), currentClip, 1.0 + currentLean) * (amplitude * fade);
    }
    return theBuffer;
}

function umtGenerateANoteSquareWave(frequency, duration, amplitude) {
    "use strict";
    // consult https://dvcs.w3.org/hg/audio/raw-file/tip/webaudio/specification.html
    var numSamples, theBuffer, bufData, i, fade;
    numSamples = duration * gUmt.globalCtx.sampleRate;
    theBuffer = gUmt.globalCtx.createBuffer(1, numSamples, gUmt.globalCtx.sampleRate); // numberOfChannels, length, sampleRate
    bufData = theBuffer.getChannelData(0);
    for (i = 0; i < numSamples; i = i + 1) {
        fade = 1.0 - (i / numSamples);
        bufData[i] = umtFMSquare(i * (frequency / gUmt.globalCtx.sampleRate)) * (amplitude * fade);
    }
    return theBuffer;
}

function umtGenerateANoteSquareVar(frequency, duration, amplitude, threshold) {
    "use strict";
    // consult https://dvcs.w3.org/hg/audio/raw-file/tip/webaudio/specification.html
    var numSamples, theBuffer, bufData, i, fade;
    threshold = (threshold / 2.0) + 0.5;
    numSamples = duration * gUmt.globalCtx.sampleRate;
    theBuffer = gUmt.globalCtx.createBuffer(1, numSamples, gUmt.globalCtx.sampleRate); // numberOfChannels, length, sampleRate
    bufData = theBuffer.getChannelData(0);
    for (i = 0; i < numSamples; i = i + 1) {
        fade = 1.0 - (i / numSamples);
        bufData[i] = umtFMVarSquare(i * (frequency / gUmt.globalCtx.sampleRate), threshold) * (amplitude * fade);
    }
    return theBuffer;
}

function umtGenerateANoteSquareRisingPitch(frequency, duration, amplitude, threshold, riseRate) {
    "use strict";
    // consult https://dvcs.w3.org/hg/audio/raw-file/tip/webaudio/specification.html
    var numSamples, theBuffer, bufData, i, fade, pctThere, expAdjustedFreq;
    threshold = (threshold / 2.0) + 0.5;
    numSamples = duration * gUmt.globalCtx.sampleRate;
    theBuffer = gUmt.globalCtx.createBuffer(1, numSamples, gUmt.globalCtx.sampleRate); // numberOfChannels, length, sampleRate
    bufData = theBuffer.getChannelData(0);
    for (i = 0; i < numSamples; i = i + 1) {
        pctThere = i / numSamples;
        fade = 1.0 - pctThere;
        expAdjustedFreq = frequency * (((i / gUmt.globalCtx.sampleRate)  * riseRate) + 1);
        bufData[i] = umtFMVarSquare(i * (expAdjustedFreq / gUmt.globalCtx.sampleRate), threshold) * (amplitude * fade);
    }
    return theBuffer;
}

function umtGenerateANoteSquareStartStopVar(frequency, duration, amplitude, startThreshold, stopThreshold) {
    "use strict";
    // consult https://dvcs.w3.org/hg/audio/raw-file/tip/webaudio/specification.html
    var numSamples, theBuffer, bufData, i, fade, thresGap, pctThere, currentThreshold;
    startThreshold = (startThreshold / 2.0) + 0.5;
    stopThreshold = (stopThreshold / 2.0) + 0.5;
    numSamples = duration * gUmt.globalCtx.sampleRate;
    theBuffer = gUmt.globalCtx.createBuffer(1, numSamples, gUmt.globalCtx.sampleRate); // numberOfChannels, length, sampleRate
    bufData = theBuffer.getChannelData(0);
    thresGap = stopThreshold - startThreshold;
    for (i = 0; i < numSamples; i = i + 1) {
        pctThere = i / numSamples;
        fade = 1.0 - pctThere;
        currentThreshold = (thresGap * pctThere) + startThreshold;
        bufData[i] = umtFMVarSquare(i * (frequency / gUmt.globalCtx.sampleRate), currentThreshold) * (amplitude * fade);
    }
    return theBuffer;
}

function umtGenerateANoteTriangleWave(frequency, duration, amplitude) {
    "use strict";
    // consult https://dvcs.w3.org/hg/audio/raw-file/tip/webaudio/specification.html
    var numSamples, theBuffer, bufData, i, fade;
    numSamples = duration * gUmt.globalCtx.sampleRate;
    theBuffer = gUmt.globalCtx.createBuffer(1, numSamples, gUmt.globalCtx.sampleRate); // numberOfChannels, length, sampleRate
    bufData = theBuffer.getChannelData(0);
    for (i = 0; i < numSamples; i = i + 1) {
        fade = 1.0 - (i / numSamples);
        bufData[i] = umtFMTriangle(i * (frequency / gUmt.globalCtx.sampleRate)) * (amplitude * fade);
    }
    return theBuffer;
}

function umtGenerateANoteTriangleVar(frequency, duration, amplitude, variance) {
    "use strict";
    // consult https://dvcs.w3.org/hg/audio/raw-file/tip/webaudio/specification.html
    var numSamples, theBuffer, bufData, i, fade;
    numSamples = duration * gUmt.globalCtx.sampleRate;
    theBuffer = gUmt.globalCtx.createBuffer(1, numSamples, gUmt.globalCtx.sampleRate); // numberOfChannels, length, sampleRate
    bufData = theBuffer.getChannelData(0);
    variance = variance + 0.5; // move origin to center point
    if (variance > 1.0) {
        variance = variance - 1.0;
    }
    for (i = 0; i < numSamples; i = i + 1) {
        fade = 1.0 - (i / numSamples);
        bufData[i] = umtFMVarTriangle(i * (frequency / gUmt.globalCtx.sampleRate), variance) * (amplitude * fade);
    }
    return theBuffer;
}

function umtGenerateANoteTriangleStartStop(frequency, duration, amplitude, startvariance, stopvariance) {
    "use strict";
    // consult https://dvcs.w3.org/hg/audio/raw-file/tip/webaudio/specification.html
    var numSamples, theBuffer, bufData, i, fade, pctThere, varGap, currentVariance;
    numSamples = duration * gUmt.globalCtx.sampleRate;
    theBuffer = gUmt.globalCtx.createBuffer(1, numSamples, gUmt.globalCtx.sampleRate); // numberOfChannels, length, sampleRate
    bufData = theBuffer.getChannelData(0);
    startvariance = startvariance + 0.5; // move origin to center point
    if (startvariance > 1.0) {
        startvariance = startvariance - 1.0;
    }
    stopvariance = stopvariance + 0.5; // move origin to center point
    if (stopvariance > 1.0) {
        stopvariance = stopvariance - 1.0;
    }
    varGap = stopvariance - startvariance;
    for (i = 0; i < numSamples; i = i + 1) {
        pctThere = i / numSamples;
        fade = 1.0 - pctThere;
        currentVariance = (varGap * pctThere) + startvariance;
        fade = 1.0 - (i / numSamples);
        bufData[i] = umtFMVarTriangle(i * (frequency / gUmt.globalCtx.sampleRate), currentVariance) * (amplitude * fade);
    }
    return theBuffer;
}

function umtGenerateANoteSawtoothWave(frequency, duration, amplitude) {
    "use strict";
    // consult https://dvcs.w3.org/hg/audio/raw-file/tip/webaudio/specification.html
    var numSamples, theBuffer, bufData, i, fade;
    numSamples = duration * gUmt.globalCtx.sampleRate;
    theBuffer = gUmt.globalCtx.createBuffer(1, numSamples, gUmt.globalCtx.sampleRate); // numberOfChannels, length, sampleRate
    bufData = theBuffer.getChannelData(0);
    for (i = 0; i < numSamples; i = i + 1) {
        fade = 1.0 - (i / numSamples);
        bufData[i] = umtFMSawtooth(i * (frequency / gUmt.globalCtx.sampleRate)) * (amplitude * fade);
    }
    return theBuffer;
}

function umtGenerateANoteSawtoothVar(frequency, duration, amplitude, variance) {
    "use strict";
    // consult https://dvcs.w3.org/hg/audio/raw-file/tip/webaudio/specification.html
    var numSamples, theBuffer, bufData, i, fade;
    numSamples = duration * gUmt.globalCtx.sampleRate;
    theBuffer = gUmt.globalCtx.createBuffer(1, numSamples, gUmt.globalCtx.sampleRate); // numberOfChannels, length, sampleRate
    bufData = theBuffer.getChannelData(0);
    // same as triange wave, except we DON'T move the origin to the center point, so we get sawtooth waves by default
    variance = variance / 5.0; // re-size to get more sawtooth wave
    for (i = 0; i < numSamples; i = i + 1) {
        fade = 1.0 - (i / numSamples);
        bufData[i] = umtFMVarTriangle(i * (frequency / gUmt.globalCtx.sampleRate), variance) * (amplitude * fade);
    }
    return theBuffer;
}

function umtGenerateANoteSawtoothStartStop(frequency, duration, amplitude, startvariance, stopvariance, riserate) {
    "use strict";
    // consult https://dvcs.w3.org/hg/audio/raw-file/tip/webaudio/specification.html
    var numSamples, theBuffer, bufData, i, fade, pctThere, varGap, currentVariance, adjustedFreq;
    numSamples = duration * gUmt.globalCtx.sampleRate;
    theBuffer = gUmt.globalCtx.createBuffer(1, numSamples, gUmt.globalCtx.sampleRate); // numberOfChannels, length, sampleRate
    bufData = theBuffer.getChannelData(0);
    // same as triange wave, except we DON'T move the origin to the center point, so we get sawtooth waves by default
    startvariance = startvariance / 5.0; // re-size to get more sawtooth wave
    stopvariance = stopvariance / 5.0;
    varGap = stopvariance - startvariance;
    adjustedFreq = frequency;
    for (i = 0; i < numSamples; i = i + 1) {
        pctThere = i / numSamples;
        fade = 1.0 - pctThere;
        currentVariance = (varGap * pctThere) + startvariance;
        fade = 1.0 - (i / numSamples);
        if (riserate !== 0) {
            adjustedFreq = frequency * (((i / gUmt.globalCtx.sampleRate)  * riserate) + 1);
        }
        bufData[i] = umtFMVarTriangle(i * (adjustedFreq / gUmt.globalCtx.sampleRate), currentVariance) * (amplitude * fade);
    }
    return theBuffer;
}

function umtGenerateANoteNoise(frequency, duration, amplitude) {
    "use strict";
    // consult https://dvcs.w3.org/hg/audio/raw-file/tip/webaudio/specification.html
    var numSamples, theBuffer, bufData, rng, i, fade, noiseSample;
    noiseSample = frequency; // delete me -- just to pass camelCaseJSLint
    numSamples = duration * gUmt.globalCtx.sampleRate;
    theBuffer = gUmt.globalCtx.createBuffer(1, numSamples, gUmt.globalCtx.sampleRate); // numberOfChannels, length, sampleRate
    bufData = theBuffer.getChannelData(0);
    rng = umtGetRando(0);
    for (i = 0; i < numSamples; i = i + 1) {
        fade = 1.0 - (i / numSamples);
        noiseSample = (rng.genrandReal3() * 2.0) - 1.0;
        bufData[i] = noiseSample * (amplitude * fade);
    }
    return theBuffer;
}

function umtGenerateANoteBackwardsNoise(frequency, duration, amplitude) {
    "use strict";
    // consult https://dvcs.w3.org/hg/audio/raw-file/tip/webaudio/specification.html
    var numSamples, theBuffer, bufData, rng, i, fade, noiseSample;
    noiseSample = frequency; // delete me -- just to pass camelCaseJSLint
    numSamples = duration * gUmt.globalCtx.sampleRate;
    theBuffer = gUmt.globalCtx.createBuffer(1, numSamples, gUmt.globalCtx.sampleRate); // numberOfChannels, length, sampleRate
    bufData = theBuffer.getChannelData(0);
    rng = umtGetRando(0);
    for (i = 0; i < numSamples; i = i + 1) {
        fade = i / numSamples;
        noiseSample = (rng.genrandReal3() * 2.0) - 1.0;
        bufData[i] = noiseSample * (amplitude * fade);
    }
    return theBuffer;
}

function umtGenerateANoteSteamPop(frequency, duration, amplitude) {
    "use strict";
    // consult https://dvcs.w3.org/hg/audio/raw-file/tip/webaudio/specification.html
    var numSamples, theBuffer, bufData, rng, i, fadeLin, fadeRecip, fadeActual, noiseSample;
    noiseSample = frequency; // delete me -- just to pass camelCaseJSLint
    numSamples = duration * gUmt.globalCtx.sampleRate;
    theBuffer = gUmt.globalCtx.createBuffer(1, numSamples, gUmt.globalCtx.sampleRate); // numberOfChannels, length, sampleRate
    bufData = theBuffer.getChannelData(0);
    rng = umtGetRando(0);
    for (i = 0; i < numSamples; i = i + 1) {
        fadeLin = 1.0 - (i / numSamples);
        // fadeRecip = 1.0 / (100.0 * (i / numSamples) + 1.0);
        fadeRecip = 1.0 / ((i / 20.0) + 1.0);
        if (fadeLin < fadeRecip) {
            fadeActual = fadeLin;
        } else {
            fadeActual = fadeRecip;
        }
        noiseSample = (rng.genrandReal3() * 2.0) - 1.0;
        bufData[i] = noiseSample * (amplitude * fadeActual);
    }
    return theBuffer;
}

function umtGenerateANoteFMSynthOriginal(frequency, duration, amplitude) {
    "use strict";
    // consult https://dvcs.w3.org/hg/audio/raw-file/tip/webaudio/specification.html
    var numSamples, theBuffer, bufData, i, fade, sinePart, subFrequency, modulatorFreq, beta;
    numSamples = duration * gUmt.globalCtx.sampleRate;
    theBuffer = gUmt.globalCtx.createBuffer(1, numSamples, gUmt.globalCtx.sampleRate); // numberOfChannels, length, sampleRate
    bufData = theBuffer.getChannelData(0);
    modulatorFreq = frequency * (3.0 / 2.0);
    beta = 10.0;
    for (i = 0; i < numSamples; i = i + 1) {
        fade = 1.0 - (i / numSamples);
        subFrequency = frequency + beta * ((Math.sin((gUmt.TAU * i * (modulatorFreq / gUmt.globalCtx.sampleRate)))));
        sinePart = Math.sin((gUmt.TAU * i * (subFrequency / gUmt.globalCtx.sampleRate)));
        bufData[i] = sinePart * (amplitude * fade);
    }
    return theBuffer;
}

function umtGenerateANoteFMSynthVar(frequency, duration, amplitude, modulator, beta) {
    "use strict";
    // consult https://dvcs.w3.org/hg/audio/raw-file/tip/webaudio/specification.html
    var numSamples, theBuffer, bufData, i, fade, sinePart, subFrequency, modList, modNum, modDnom, modulatorFreq;
    numSamples = duration * gUmt.globalCtx.sampleRate;
    theBuffer = gUmt.globalCtx.createBuffer(1, numSamples, gUmt.globalCtx.sampleRate); // numberOfChannels, length, sampleRate
    bufData = theBuffer.getChannelData(0);
    modList = [ {num: 3, dnom: 2}, {num: 4, dnom: 3}, {num: 5, dnom: 3}, {num: 6, dnom: 5}, {num: 5, dnom: 4}, {num: 8, dnom: 5}, {num: 9, dnom: 5}, {num: 10, dnom: 9}, {num: 9, dnom: 8}, {num: 16, dnom: 9}, {num: 16, dnom: 15}, {num: 15, dnom: 8} ];
    modulator = Math.floor(modulator * (modList.length - 1));
    modNum = modList[modulator].num;
    modDnom = modList[modulator].dnom;
    modulatorFreq = frequency * (modNum / modDnom);
    beta = beta * Math.exp((beta - 0.7) * 10); // multiplying by beta is kinda hokey but casts 0 to 0 which is what we want.
    for (i = 0; i < numSamples; i = i + 1) {
        fade = 1.0 - (i / numSamples);
        subFrequency = frequency + beta * ((Math.sin((gUmt.TAU * i * (modulatorFreq / gUmt.globalCtx.sampleRate)))));
        sinePart = Math.sin((gUmt.TAU * i * (subFrequency / gUmt.globalCtx.sampleRate)));
        bufData[i] = sinePart * (amplitude * fade);
    }
    return theBuffer;
}

function umtGenerateANoteFMSynthStartStop(frequency, duration, amplitude, startModulator, startBeta, stopModulator, stopBeta) {
    "use strict";
    // consult https://dvcs.w3.org/hg/audio/raw-file/tip/webaudio/specification.html
    var numSamples, theBuffer, bufData, i, fade, sinePart, subFrequency, modList, startModNum, startModDnom, startModulatorFreq, stopModNum, stopModDnom, stopModulatorFreq, pctThere, modulatorDiff, betaDiff, currentModulatorFreq, currentBeta;
    numSamples = duration * gUmt.globalCtx.sampleRate;
    theBuffer = gUmt.globalCtx.createBuffer(1, numSamples, gUmt.globalCtx.sampleRate); // numberOfChannels, length, sampleRate
    bufData = theBuffer.getChannelData(0);
    modList = [ {num: 3, dnom: 2}, {num: 4, dnom: 3}, {num: 5, dnom: 3}, {num: 6, dnom: 5}, {num: 5, dnom: 4}, {num: 8, dnom: 5}, {num: 9, dnom: 5}, {num: 10, dnom: 9}, {num: 9, dnom: 8}, {num: 16, dnom: 9}, {num: 16, dnom: 15}, {num: 15, dnom: 8} ];
    startModulator = Math.floor(startModulator * (modList.length - 1));
    startModNum = modList[startModulator].num;
    startModDnom = modList[startModulator].dnom;
    startModulatorFreq = frequency * (startModNum / startModDnom);
    stopModulator = Math.floor(stopModulator * (modList.length - 1));
    stopModNum = modList[stopModulator].num;
    stopModDnom = modList[stopModulator].dnom;
    stopModulatorFreq = frequency * (stopModNum / stopModDnom);
    startBeta = (startBeta * 5.0);
    stopBeta = (stopBeta * 5.0);
    modulatorDiff = stopModulatorFreq - startModulatorFreq;
    betaDiff = stopBeta - startBeta;
    for (i = 0; i < numSamples; i = i + 1) {
        pctThere = (i / numSamples);
        fade = 1.0 - pctThere;
        currentModulatorFreq = (modulatorDiff * pctThere) + startModulatorFreq;
        currentBeta = (betaDiff * pctThere) + startBeta;
        currentBeta = currentBeta * Math.exp((currentBeta - 0.7) * 10); // multiplying by beta is kinda hokey but casts 0 to 0 which is what we want.
        subFrequency = frequency + currentBeta * ((Math.sin((gUmt.TAU * i * (currentModulatorFreq / gUmt.globalCtx.sampleRate)))));
        sinePart = Math.sin((gUmt.TAU * i * (subFrequency / gUmt.globalCtx.sampleRate)));
        bufData[i] = sinePart * (amplitude * fade);
    }
    return theBuffer;
}

function umtGenerateANoteFMSynthSquare(frequency, duration, amplitude, startThreshold, stopThreshold, startModulator, startBeta, stopModulator, stopBeta) {
    "use strict";
    // consult https://dvcs.w3.org/hg/audio/raw-file/tip/webaudio/specification.html
    var numSamples, theBuffer, bufData, i, fade, squarePart, subFrequency, modList, startModNum, startModDnom, startModulatorFreq, stopModNum, stopModDnom, stopModulatorFreq, pctThere, thresDiff, modulatorDiff, betaDiff, currentThreshold, currentModulatorFreq, currentBeta;
    numSamples = duration * gUmt.globalCtx.sampleRate;
    theBuffer = gUmt.globalCtx.createBuffer(1, numSamples, gUmt.globalCtx.sampleRate); // numberOfChannels, length, sampleRate
    bufData = theBuffer.getChannelData(0);
    startThreshold = (startThreshold / 2.0) + 0.5;
    stopThreshold = (stopThreshold / 2.0) + 0.5;
    modList = [ {num: 3, dnom: 2}, {num: 4, dnom: 3}, {num: 5, dnom: 3}, {num: 6, dnom: 5}, {num: 5, dnom: 4}, {num: 8, dnom: 5}, {num: 9, dnom: 5}, {num: 10, dnom: 9}, {num: 9, dnom: 8}, {num: 16, dnom: 9}, {num: 16, dnom: 15}, {num: 15, dnom: 8} ];
    startModulator = Math.floor(startModulator * (modList.length - 1));
    startModNum = modList[startModulator].num;
    startModDnom = modList[startModulator].dnom;
    startModulatorFreq = frequency * (startModNum / startModDnom);
    stopModulator = Math.floor(stopModulator * (modList.length - 1));
    stopModNum = modList[stopModulator].num;
    stopModDnom = modList[stopModulator].dnom;
    stopModulatorFreq = frequency * (stopModNum / stopModDnom);
    thresDiff = stopThreshold - startThreshold;
    modulatorDiff = stopModulatorFreq - startModulatorFreq;
    betaDiff = stopBeta - startBeta;
    for (i = 0; i < numSamples; i = i + 1) {
        pctThere = (i / numSamples);
        fade = 1.0 - pctThere;
        currentModulatorFreq = (modulatorDiff * pctThere) + startModulatorFreq;
        currentBeta = (betaDiff * pctThere) + startBeta;
        currentBeta = currentBeta * Math.exp((currentBeta - 0.5) * 10); // multiplying by beta is kinda hokey but casts 0 to 0 which is what we want.
        subFrequency = frequency + currentBeta * ((Math.sin((gUmt.TAU * i * (currentModulatorFreq / gUmt.globalCtx.sampleRate)))));
        // sinePart = Math.sin((gUmt.TAU * i * (subFrequency / gUmt.globalCtx.sampleRate)));
        // squarePart = umtFMSquare(i * (subFrequency / gUmt.globalCtx.sampleRate));
        currentThreshold = (thresDiff * pctThere) + startThreshold;
        squarePart = umtFMVarSquare(i * (subFrequency / gUmt.globalCtx.sampleRate), currentThreshold);
        bufData[i] = squarePart * (amplitude * fade);
    }
    return theBuffer;
}

function umtGenerateANoteFMSynthCrossNote(startMoment, frequency, duration, amplitude, modulator, beta) {
    "use strict";
    // consult https://dvcs.w3.org/hg/audio/raw-file/tip/webaudio/specification.html
    var numSamples, theBuffer, bufData, i, fade, sinePart, subFrequency, modulatorFreq, offset;
    numSamples = duration * gUmt.globalCtx.sampleRate;
    theBuffer = gUmt.globalCtx.createBuffer(1, numSamples, gUmt.globalCtx.sampleRate); // numberOfChannels, length, sampleRate
    bufData = theBuffer.getChannelData(0);
    modulatorFreq = modulator;
    modulatorFreq = frequency * (2.0 / 3.0);
    offset = (startMoment - (Math.floor(startMoment / modulatorFreq) * modulatorFreq)) * gUmt.globalCtx.sampleRate;
    beta = beta * Math.exp((beta - 0.7) * 10); // multiplying by beta is kinda hokey but casts 0 to 0 which is what we want.
    for (i = 0; i < numSamples; i = i + 1) {
        fade = 1.0; //  - (i / numSamples);
        subFrequency = frequency + beta * ((Math.sin((gUmt.TAU * (i + offset) * (modulatorFreq / gUmt.globalCtx.sampleRate)))));
        sinePart = Math.sin((gUmt.TAU * (i + offset) * (subFrequency / gUmt.globalCtx.sampleRate)));
        bufData[i] = sinePart * (amplitude * fade);
    }
    return theBuffer;
}

function InstantiateTuningForkObj() {
    "use strict";
    this.getParams = function () {
        return {
            percussion: false,
            fixed: [],
            parameters: [
                { name: "clipping", display: "Clipping" },
                { name: "lean", display: "Lean" }
            ]
        };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var idxname, theBuffer, node, clipping, lean;
        clipping = instSpecificParams.clipping;
        lean = instSpecificParams.lean;
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = "sinewave" + frequency + "x" + duration + "x" + amplitude + "x" + clipping + "x" + lean;
        if (gUmt.cachedNotes.hasOwnProperty(idxname)) {
            theBuffer = gUmt.cachedNotes[idxname];
        } else {
            theBuffer = umtGenerateANoteSineWave(frequency, duration, amplitude, clipping, lean);
            gUmt.cachedNotes[idxname] = theBuffer;
        }
        node = gUmt.globalCtx.createBufferSource();
        node.buffer = theBuffer;
        node.connect(gUmt.globalCtx.destination);
        node.start(startMoment);
    };
}

function InstantiateSineStartStopObj() {
    "use strict";
    this.getParams = function () {
        return {
            percussion: false,
            fixed: [],
            parameters: [
                { name: "startclipping", display: "Clipping start" },
                { name: "startlean", display: "Lean start" },
                { name: "stopclipping", display: "Clipping stop" },
                { name: "stoplean", display: "Lean stop" }
            ]
        };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var idxname, theBuffer, node, startclipping, startlean, stopclipping, stoplean;
        startclipping = instSpecificParams.startclipping;
        startlean = instSpecificParams.startlean;
        stopclipping = instSpecificParams.stopclipping;
        stoplean = instSpecificParams.stoplean;
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = "sinewave" + frequency + "x" + duration + "x" + amplitude + "x" + startclipping + "x" + startlean + "x" + stopclipping + "x" + stoplean;
        if (gUmt.cachedNotes.hasOwnProperty(idxname)) {
            theBuffer = gUmt.cachedNotes[idxname];
        } else {
            theBuffer = umtGenerateANoteSineStartStopWave(frequency, duration, amplitude, startclipping, startlean, stopclipping, stoplean);
            gUmt.cachedNotes[idxname] = theBuffer;
        }
        node = gUmt.globalCtx.createBufferSource();
        node.buffer = theBuffer;
        node.connect(gUmt.globalCtx.destination);
        node.start(startMoment);
    };
}

function InstantiateSquareWaveObj() {
    "use strict";
    this.getParams = function () {
        return { percussion: false, fixed: [], parameters: [] };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var idxname, theBuffer, node;
        idxname = instSpecificParams; // jslint-temp
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = "squarewave" + frequency + "x" + duration + "x" + amplitude;
        if (gUmt.cachedNotes.hasOwnProperty(idxname)) {
            theBuffer = gUmt.cachedNotes[idxname];
        } else {
            theBuffer = umtGenerateANoteSquareWave(frequency, duration, amplitude);
            gUmt.cachedNotes[idxname] = theBuffer;
        }
        node = gUmt.globalCtx.createBufferSource();
        node.buffer = theBuffer;
        node.connect(gUmt.globalCtx.destination);
        node.start(startMoment);
    };
}

function InstantiateSquareVarObj() {
    "use strict";
    this.getParams = function () {
        return { percussion: false, fixed: [], parameters: [ { name: "unsquareness", display: "Un-square-ness" } ] };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var idxname, theBuffer, node, threshold;
        threshold = instSpecificParams.unsquareness;
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = "squarevar" + frequency + "x" + duration + "x" + amplitude + "x" + threshold;
        if (gUmt.cachedNotes.hasOwnProperty(idxname)) {
            theBuffer = gUmt.cachedNotes[idxname];
        } else {
            theBuffer = umtGenerateANoteSquareVar(frequency, duration, amplitude, threshold * 0.96);
            gUmt.cachedNotes[idxname] = theBuffer;
        }
        node = gUmt.globalCtx.createBufferSource();
        node.buffer = theBuffer;
        node.connect(gUmt.globalCtx.destination);
        node.start(startMoment);
    };
}

function InstantiateSquareRisingPitchObj() {
    "use strict";
    this.getParams = function () {
        return {
            percussion: false,
            fixed: false,
            parameters: [
                { name: "unsquareness", display: "Un-square-ness" },
                { name: "riserate", display: "Pitch rise rate" }
            ]
        };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var idxname, theBuffer, node, threshold, riserate;
        threshold = instSpecificParams.unsquareness;
        riserate = instSpecificParams.riserate;
        riserate = riserate * 4;
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = "squarerise" + frequency + "x" + duration + "x" + amplitude + "x" + threshold + "x" + riserate;
        if (gUmt.cachedNotes.hasOwnProperty(idxname)) {
            theBuffer = gUmt.cachedNotes[idxname];
        } else {
            theBuffer = umtGenerateANoteSquareRisingPitch(frequency, duration, amplitude, threshold * 0.96, riserate);
            gUmt.cachedNotes[idxname] = theBuffer;
        }
        node = gUmt.globalCtx.createBufferSource();
        node.buffer = theBuffer;
        node.connect(gUmt.globalCtx.destination);
        node.start(startMoment);
    };
}

function InstantiateSquareStartStopObj() {
    "use strict";
    this.getParams = function () {
        return {
            percussion: false,
            fixed: false,
            parameters: [
                { name: "startunsquareness", display: "Start un-square-ness" },
                { name: "stopunsquareness", display: "Stop un-square-ness" }
            ]
        };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var idxname, theBuffer, node, startThreshold, stopThreshold;
        startThreshold = instSpecificParams.startunsquareness;
        stopThreshold = instSpecificParams.stopunsquareness;
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = "squaredblvar" + frequency + "x" + duration + "x" + amplitude + "x" + startThreshold + "x" + stopThreshold;
        if (gUmt.cachedNotes.hasOwnProperty(idxname)) {
            theBuffer = gUmt.cachedNotes[idxname];
        } else {
            theBuffer = umtGenerateANoteSquareStartStopVar(frequency, duration, amplitude, startThreshold * 0.96, stopThreshold * 0.96);
            gUmt.cachedNotes[idxname] = theBuffer;
        }
        node = gUmt.globalCtx.createBufferSource();
        node.buffer = theBuffer;
        node.connect(gUmt.globalCtx.destination);
        node.start(startMoment);
    };
}

function InstantiateTriangleWaveObj() {
    "use strict";
    this.getParams = function () {
        return { percussion: false, fixed: false, parameters: [] };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var idxname, theBuffer, node;
        idxname = instSpecificParams; // jslint-temp
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = "triangle" + frequency + "x" + duration + "x" + amplitude;
        if (gUmt.cachedNotes.hasOwnProperty(idxname)) {
            theBuffer = gUmt.cachedNotes[idxname];
        } else {
            theBuffer = umtGenerateANoteTriangleWave(frequency, duration, amplitude);
            gUmt.cachedNotes[idxname] = theBuffer;
        }
        node = gUmt.globalCtx.createBufferSource();
        node.buffer = theBuffer;
        node.connect(gUmt.globalCtx.destination);
        node.start(startMoment);
    };
}

function InstantiateTriangleVarObj() {
    "use strict";
    this.getParams = function () {
        return { percussion: false, fixed: [], parameters: [ { name: "untriangleness", display: "Un-Triangleness" } ] };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var variance, idxname, theBuffer, node;
        variance = instSpecificParams.untriangleness;
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = "triangle" + frequency + "x" + duration + "x" + amplitude + "x" + variance;
        if (gUmt.cachedNotes.hasOwnProperty(idxname)) {
            theBuffer = gUmt.cachedNotes[idxname];
        } else {
            theBuffer = umtGenerateANoteTriangleVar(frequency, duration, amplitude, variance);
            gUmt.cachedNotes[idxname] = theBuffer;
        }
        node = gUmt.globalCtx.createBufferSource();
        node.buffer = theBuffer;
        node.connect(gUmt.globalCtx.destination);
        node.start(startMoment);
    };
}

function InstantiateTriangleStartStopObj() {
    "use strict";
    this.getParams = function () {
        return {
            percussion: false,
            fixed: [],
            parameters: [
                { name: "startuntriangleness", display: "Un-Triangleness start" },
                { name: "stopuntriangleness", display: "Un-Triangleness stop" }
            ]
        };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var startvariance, stopvariance, idxname, theBuffer, node;
        startvariance = instSpecificParams.startuntriangleness;
        stopvariance = instSpecificParams.stopuntriangleness;
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = "triangle" + frequency + "x" + duration + "x" + amplitude + "x" + startvariance + "x" + stopvariance;
        if (gUmt.cachedNotes.hasOwnProperty(idxname)) {
            theBuffer = gUmt.cachedNotes[idxname];
        } else {
            theBuffer = umtGenerateANoteTriangleStartStop(frequency, duration, amplitude, startvariance, stopvariance);
            gUmt.cachedNotes[idxname] = theBuffer;
        }
        node = gUmt.globalCtx.createBufferSource();
        node.buffer = theBuffer;
        node.connect(gUmt.globalCtx.destination);
        node.start(startMoment);
    };
}

function InstantiateSawtoothWaveObj() {
    "use strict";
    this.getParams = function () {
        return { percussion: false, fixed: [], parameters: [] };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var idxname, theBuffer, node;
        idxname = instSpecificParams; // jslint-temp
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = "sawtooth" + frequency + "x" + duration + "x" + amplitude;
        if (gUmt.cachedNotes.hasOwnProperty(idxname)) {
            theBuffer = gUmt.cachedNotes[idxname];
        } else {
            theBuffer = umtGenerateANoteSawtoothWave(frequency, duration, amplitude);
            gUmt.cachedNotes[idxname] = theBuffer;
        }
        node = gUmt.globalCtx.createBufferSource();
        node.buffer = theBuffer;
        node.connect(gUmt.globalCtx.destination);
        node.start(startMoment);
    };
}

function InstantiateSawtoothVarObj() {
    "use strict";
    this.getParams = function () {
        return { percussion: false, fixed: [], parameters: [ { name: "unsawtoothness", display: "Un-Sawtoothness" } ] };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var variance, idxname, theBuffer, node;
        variance = instSpecificParams.unsawtoothness;
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = "sawtoothvar" + frequency + "x" + duration + "x" + amplitude + "x" + variance;
        if (gUmt.cachedNotes.hasOwnProperty(idxname)) {
            theBuffer = gUmt.cachedNotes[idxname];
        } else {
            theBuffer = umtGenerateANoteSawtoothVar(frequency, duration, amplitude, variance);
            gUmt.cachedNotes[idxname] = theBuffer;
        }
        node = gUmt.globalCtx.createBufferSource();
        node.buffer = theBuffer;
        node.connect(gUmt.globalCtx.destination);
        node.start(startMoment);
    };
}

function InstantiateSawtoothStartStopObj() {
    "use strict";
    this.getParams = function () {
        return {
            percussion: false,
            fixed: [],
            parameters: [
                { name: "startunsawtoothness", display: "Un-Sawtoothness start" },
                { name: "stopunsawtoothness", display: "Un-Sawtoothness stop" },
                { name: "riserate", display: "Rise rate" }
            ]
        };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var startvariance, stopvariance, riserate, idxname, theBuffer, node;
        startvariance = instSpecificParams.startunsawtoothness;
        stopvariance = instSpecificParams.stopunsawtoothness;
        riserate = instSpecificParams.riserate;
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = "sawtoothvar" + frequency + "x" + duration + "x" + amplitude + "x" + startvariance + "x" + stopvariance + "x" + riserate;
        if (gUmt.cachedNotes.hasOwnProperty(idxname)) {
            theBuffer = gUmt.cachedNotes[idxname];
        } else {
            theBuffer = umtGenerateANoteSawtoothStartStop(frequency, duration, amplitude, startvariance, stopvariance, riserate);
            gUmt.cachedNotes[idxname] = theBuffer;
        }
        node = gUmt.globalCtx.createBufferSource();
        node.buffer = theBuffer;
        node.connect(gUmt.globalCtx.destination);
        node.start(startMoment);
    };
}

function InstantiateNoiseObj() {
    "use strict";
    this.getParams = function () {
        return { percussion: true, fixed: [], parameters: [] };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var idxname, theBuffer, node;
        idxname = instSpecificParams; // jslint-temp
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = "forwardsnoise" + frequency + "x" + duration + "x" + amplitude;
        if (gUmt.cachedNotes.hasOwnProperty(idxname)) {
            theBuffer = gUmt.cachedNotes[idxname];
        } else {
            theBuffer = umtGenerateANoteNoise(frequency, duration, amplitude);
            gUmt.cachedNotes[idxname] = theBuffer;
        }
        node = gUmt.globalCtx.createBufferSource();
        node.buffer = theBuffer;
        node.connect(gUmt.globalCtx.destination);
        node.start(startMoment);
    };
}

function InstantiateBackwardsNoiseObj() {
    "use strict";
    this.getParams = function () {
        return { percussion: true, fixed: [], parameters: [] };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var idxname, theBuffer, node;
        idxname = instSpecificParams; // jslint-temp
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = "backwardsnoise" + frequency + "x" + duration + "x" + amplitude;
        if (gUmt.cachedNotes.hasOwnProperty(idxname)) {
            theBuffer = gUmt.cachedNotes[idxname];
        } else {
            theBuffer = umtGenerateANoteBackwardsNoise(frequency, duration, amplitude);
            gUmt.cachedNotes[idxname] = theBuffer;
        }
        node = gUmt.globalCtx.createBufferSource();
        node.buffer = theBuffer;
        node.connect(gUmt.globalCtx.destination);
        node.start(startMoment);
    };
}

function InstantiateSteamPopObj() {
    "use strict";
    this.getParams = function () {
        return { percussion: true, fixed: [], parameters: [] };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var idxname, theBuffer, node;
        idxname = instSpecificParams; // jslint-temp
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = "steampop" + frequency + "x" + duration + "x" + amplitude;
        if (gUmt.cachedNotes.hasOwnProperty(idxname)) {
            theBuffer = gUmt.cachedNotes[idxname];
        } else {
            theBuffer = umtGenerateANoteSteamPop(frequency, duration, amplitude);
            gUmt.cachedNotes[idxname] = theBuffer;
        }
        node = gUmt.globalCtx.createBufferSource();
        node.buffer = theBuffer;
        node.connect(gUmt.globalCtx.destination);
        node.start(startMoment);
    };
}

function InstantiateFMSynthOriginalObj() {
    "use strict";
    this.getParams = function () {
        return { percussion: false, fixed: [], parameters: [] };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var idxname, theBuffer, node;
        idxname = instSpecificParams; // jslint-temp
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = "fmsynthoriginal" + frequency + "x" + duration + "x" + amplitude;
        if (gUmt.cachedNotes.hasOwnProperty(idxname)) {
            theBuffer = gUmt.cachedNotes[idxname];
        } else {
            theBuffer = umtGenerateANoteFMSynthOriginal(frequency, duration, amplitude);
            gUmt.cachedNotes[idxname] = theBuffer;
        }
        node = gUmt.globalCtx.createBufferSource();
        node.buffer = theBuffer;
        node.connect(gUmt.globalCtx.destination);
        node.start(startMoment);
    };
}

function InstantiateFMSynthVarObj() {
    "use strict";
    this.getParams = function () {
        return { percussion: false, fixed: [], parameters: [
            { name: "modulator", display: "Modulator" },
            { name: "beta", display: "Beta" }
        ] };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var modulator, beta, idxname, theBuffer, node;
        modulator = instSpecificParams.modulator;
        beta = instSpecificParams.beta;
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = "fmsynthvar" + frequency + "x" + duration + "x" + amplitude + "x" + modulator + "x" + beta;
        if (gUmt.cachedNotes.hasOwnProperty(idxname)) {
            theBuffer = gUmt.cachedNotes[idxname];
        } else {
            theBuffer = umtGenerateANoteFMSynthVar(frequency, duration, amplitude, modulator, beta);
            gUmt.cachedNotes[idxname] = theBuffer;
        }
        node = gUmt.globalCtx.createBufferSource();
        node.buffer = theBuffer;
        node.connect(gUmt.globalCtx.destination);
        node.start(startMoment);
    };
}

function InstantiateFMSynthStartStopObj() {
    "use strict";
    this.getParams = function () {
        return { percussion: false, fixed: [], parameters: [
            { name: "startmodulator", display: "Start modulator" },
            { name: "startbeta", display: "Start beta" },
            { name: "stopmodulator", display: "Stop modulator" },
            { name: "stopbeta", display: "Stop beta" }
        ] };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var startmodulator, startbeta, stopmodulator, stopbeta, idxname, theBuffer, node;
        startmodulator = instSpecificParams.startmodulator;
        startbeta = instSpecificParams.startbeta;
        stopmodulator = instSpecificParams.stopmodulator;
        stopbeta = instSpecificParams.stopbeta;
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = "fmsynthstartstop" + frequency + "x" + duration + "x" + amplitude + "x" + startmodulator + "x" + startbeta + "x" + stopmodulator + "x" + stopbeta;
        if (gUmt.cachedNotes.hasOwnProperty(idxname)) {
            theBuffer = gUmt.cachedNotes[idxname];
        } else {
            theBuffer = umtGenerateANoteFMSynthStartStop(frequency, duration, amplitude, startmodulator, startbeta, stopmodulator, stopbeta);
            gUmt.cachedNotes[idxname] = theBuffer;
        }
        node = gUmt.globalCtx.createBufferSource();
        node.buffer = theBuffer;
        node.connect(gUmt.globalCtx.destination);
        node.start(startMoment);
    };
}

function InstantiateFMSynthSquareObj() {
    "use strict";
    this.getParams = function () {
        return {
            percussion: false,
            fixed: [],
            parameters: [
                { name: "startunsquareness", display: "Start un-square-ness" },
                { name: "startmodulator", display: "Start modulator" },
                { name: "startbeta", display: "Start beta" },
                { name: "stopunsquareness", display: "Stop un-square-ness" },
                { name: "stopmodulator", display: "Stop modulator" },
                { name: "stopbeta", display: "Stop beta" }
            ]
        };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var startThreshold, stopThreshold, startModulator, startBeta, stopModulator, stopBeta, idxname, theBuffer, node;
        startThreshold = instSpecificParams.startunsquareness;
        stopThreshold = instSpecificParams.stopunsquareness;
        startModulator = instSpecificParams.startmodulator;
        startBeta = instSpecificParams.startbeta;
        stopModulator = instSpecificParams.stopmodulator;
        stopBeta = instSpecificParams.stopbeta;
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = "fmstartstpsquare" + frequency + "x" + duration + "x" + amplitude + "x" + startThreshold + "X" + stopThreshold + "x" + startModulator + "x" + startBeta + "x" + stopModulator + "x" + stopBeta;
        if (gUmt.cachedNotes.hasOwnProperty(idxname)) {
            theBuffer = gUmt.cachedNotes[idxname];
        } else {
            theBuffer = umtGenerateANoteFMSynthSquare(frequency, duration, amplitude, startThreshold * 0.96, stopThreshold * 0.96, startModulator, startBeta, stopModulator, stopBeta);
            gUmt.cachedNotes[idxname] = theBuffer;
        }
        node = gUmt.globalCtx.createBufferSource();
        node.buffer = theBuffer;
        node.connect(gUmt.globalCtx.destination);
        node.start(startMoment);
        console.log(originalNote); // jslint-temp
        console.log(originalStartTime); // jslint-temp
    };
}

function InstantiateFMSynthCrossNoteObj() {
    "use strict";
    this.getParams = function () {
        return {
            percussion: false,
            fixed: [],
            parameters: [
                { name: "modulator", display: "Modulator" },
                { name: "beta", display: "Beta" }
            ]
        };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var modulator, beta, theBuffer, node;
        modulator = instSpecificParams.modulator;
        beta = instSpecificParams.beta;
        theBuffer = umtGenerateANoteFMSynthCrossNote(startMoment, frequency, duration, amplitude, modulator, beta);
        node = gUmt.globalCtx.createBufferSource();
        node.buffer = theBuffer;
        node.connect(gUmt.globalCtx.destination);
        node.start(startMoment);
        console.log(originalNote); // jslint-temp
        console.log(originalStartTime); // jslint-temp
    };
}

function InstantiateWadSinObj() {
    "use strict";
    this.getParams = function () {
        return { percussion: false, fixed: [], parameters: [] };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var idxname, theWad;
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = instSpecificParams; // delete me -- just to pass camelCaseJSLint
        idxname = "wadsin";
        if (gUmt.cachedWads.hasOwnProperty(idxname)) {
            theWad = gUmt.cachedWads[idxname];
        } else {
            theWad = new Wad({ source: 'sine', volume: amplitude, env: { attack: 0.0, decay: 0.1, sustain: 1.0, hold: duration, release: 0 } });
            gUmt.cachedWads[idxname] = theWad;
        }
        theWad.play({ exactTime: startMoment, pitch: frequency, volume: amplitude, env: {hold: duration} });
    };
}

function InstantiateWadKickObj() {
    "use strict";
    this.getParams = function () {
        return { percussion: true, fixed: [], parameters: [] };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var idxname, theWad;
        idxname = duration; // delete me -- just to pass camelCaseJSLint
        idxname = amplitude; // delete me -- just to pass camelCaseJSLint
        idxname = instSpecificParams; // jslint-temp
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = "wadkick";
        if (gUmt.cachedWads.hasOwnProperty(idxname)) {
            theWad = gUmt.cachedWads[idxname];
        } else {
            theWad = new Wad({ source: 'http://www.codecur.io/us/sendaudio/kick.mp3' });
            gUmt.cachedWads[idxname] = theWad;
        }
        theWad.play({ exactTime: startMoment, pitch: frequency });
    };
}

function InstantiateWadBassObj() {
    "use strict";
    this.getParams = function () {
        return { percussion: false, fixed: [], parameters: [] };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var idxname, theWad;
        idxname = instSpecificParams; // jslint-temp
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = "wadbass";
        if (gUmt.cachedWads.hasOwnProperty(idxname)) {
            theWad = gUmt.cachedWads[idxname];
        } else {
            theWad = new Wad({ source: 'sine', volume: amplitude, globalReverb: true, env: { attack: 0.02, decay: 0.1, sustain: 0.9, hold: duration, release: 0.1 } });
            gUmt.cachedWads[idxname] = theWad;
        }
        theWad.play({ exactTime: startMoment, pitch: frequency, volume: amplitude, env: {hold: duration} });
    };
}

function InstantiateWadSnareObj() {
    "use strict";
    this.getParams = function () {
        return { percussion: true, fixed: [], parameters: [] };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var idxname, theWad;
        idxname = instSpecificParams; // jslint-temp
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = "wadsnare";
        if (gUmt.cachedWads.hasOwnProperty(idxname)) {
            theWad = gUmt.cachedWads[idxname];
        } else {
            theWad = new Wad(Wad.presets.snare);
            theWad.globalReverb = true;
            gUmt.cachedWads[idxname] = theWad;
        }
        theWad.play({ exactTime: startMoment, pitch: frequency, volume: amplitude, env: {hold: duration} });
    };
}

function InstantiateWadHihatOpenObj() {
    "use strict";
    this.getParams = function () {
        return { percussion: false, fixed: [], parameters: [] };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var idxname, theWad;
        idxname = duration; // delete me -- just to pass camelCaseJSLint
        idxname = instSpecificParams; // delete me -- just to pass camelCaseJSLint
        // idxname = "wadhihats" + frequency + "x" + duration + "x" + amplitude;
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        idxname = "wadhihatopen";
        if (gUmt.cachedWads.hasOwnProperty(idxname)) {
            theWad = gUmt.cachedWads[idxname];
        } else {
            theWad = new Wad(Wad.presets.hiHatOpen);
            theWad.globalReverb = true;
            gUmt.cachedWads[idxname] = theWad;
        }
        theWad.play({ exactTime: startMoment, pitch: frequency, volume: amplitude });
    };
}

function InstantiateWadHihatClosedObj() {
    "use strict";
    this.getParams = function () {
        return { percussion: false, fixed: [], parameters: [] };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var idxname, theWad;
        idxname = duration; // delete me -- just to pass camelCaseJSLint
        idxname = instSpecificParams; // jslint-temp
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        // idxname = "wadhihats" + frequency + "x" + duration + "x" + amplitude;
        idxname = "wadhihatclosed";
        if (gUmt.cachedWads.hasOwnProperty(idxname)) {
            theWad = gUmt.cachedWads[idxname];
        } else {
            theWad = new Wad(Wad.presets.hiHatClosed);
            theWad.globalReverb = true;
            gUmt.cachedWads[idxname] = theWad;
        }
        theWad.play({ exactTime: startMoment, pitch: frequency, volume: amplitude });
    };
}

function InstantiateWadFluteObj() {
    "use strict";
    this.getParams = function () {
        return { percussion: false, fixed: [], parameters: [] };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var idxname, theWad;
        idxname = instSpecificParams; // jslint-temp
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        // idxname = "wadflute" + frequency + "x" + duration + "x" + amplitude;
        idxname = "wadflute";
        if (gUmt.cachedWads.hasOwnProperty(idxname)) {
            theWad = gUmt.cachedWads[idxname];
        } else {
            theWad = new Wad(Wad.presets.ghost);
            // Wad.setGlobalReverb({ impulse: 'http://www.codecur.io/us/sendaudio/longhall.wav', wet : 0.5 });
            gUmt.cachedWads[idxname] = theWad;
        }
        theWad.play({ exactTime: startMoment, pitch: frequency, volume: amplitude, env: {hold: duration} });
    };
}

function InstantiateWadPianoObj() {
    "use strict";
    this.getParams = function () {
        return { percussion: false, fixed: [], parameters: [] };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var idxname, theWad;
        idxname = instSpecificParams; // jslint-temp
        idxname = originalNote + "z" + originalStartTime; // jslint-temp
        // idxname = "wadpiano" + frequency + "x" + duration + "x" + amplitude;
        idxname = "wadpiano";
        if (gUmt.cachedWads.hasOwnProperty(idxname)) {
            theWad = gUmt.cachedWads[idxname];
        } else {
            theWad = new Wad({source: 'square', volume: amplitude, env: { attack: 0.01, decay: 0.005, sustain: 0.2, hold: duration, release: 0.3}, filter: { type: 'lowpass', frequency: 1200, q: 8.5, env: { attack: 0.2, frequency: 600 } } });
            theWad.globalReverb = true;
            gUmt.cachedWads[idxname] = theWad;
        }
        theWad.play({ exactTime: startMoment, pitch: frequency, volume: amplitude, env: {hold: duration} });
    };
}

function InstantiateDanLights() {
    "use strict";
    this.getParams = function () {
        return {
            percussion: true,
            fixed: [
                { name: "bank", display: "Bank", type: "list", values: ["lobbywall=Lobby Wall", "lobbylanterns=Lobby Lanterns", "baywhite=Bay White", "baycolor=Bay Color"], default: "lobbywall" },
                { name: "unit", display: "Unit", type: "list", values: ["0=*", "1=1", "2=2", "3=3", "4=4", "5=5", "6=6", "7=7", "8=8", "9=9", "10=10", "11=11", "12=12"], default: "1" },
                { name: "basecolor", display: "Base Color", type: "list", values: ["red=Red", "yellow=Yellow", "green=Green", "cyan=Cyan", "blue=Blue", "magenta=Magenta"], default: "red" },
                { name: "assignedbit", display: "Assigned Bit", type: "list", values: ["0=0", "1=1", "2=2", "3=3", "4=4", "5=5", "6=6", "7=7", "8=8", "9=9", "10=10", "11=11", "12=12", "13=13", "14=14", "15=15", "16=16"], default: "0" }
            ],
            parameters: [
                { name: "pastelness", display: "Pastel-ness" },
                { name: "twist", display: "Twist" }
            ]
        };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var currentTime, startHue, stopHue, saturation, bank, unit, twist, bitnum, bitvalue, i;
        console.log("startMoment");
        console.log(startMoment);
        console.log("frequency");
        console.log(frequency);
        console.log("duration");
        console.log(duration);
        console.log("amplitude");
        console.log(amplitude);
        console.log("instSpecificParams");
        console.log(instSpecificParams);
        console.log("originalNote");
        console.log(originalNote);
        console.log("originalStartTime");
        console.log(originalStartTime);
        console.log("-----");
        if (instSpecificParams.assignedbit > 0) {
            bitnum = instSpecificParams.assignedbit;
            bitvalue = 1;
            for (i = bitnum; i > 0; i = i - 1) {
                bitvalue = bitvalue * 2;
            }
            /*jslint bitwise: true */
            if ((originalStartTime & bitvalue) === 0) {
                return;
            }
            if ((originalStartTime & (bitvalue - 1)) !== 0) {
                return;
            }
            /*jslint bitwise: false */
        }
        // time now, start time, duration, instrument, instrument specific params...
        // our inst specific params: hue, saturation
        currentTime = gUmt.globalCtx.currentTime;
        switch (instSpecificParams.basecolor) {
        case 'red':
            startHue = 0;
            break;
        case 'yellow':
            startHue = 47;
            break;
        case 'green':
            startHue = 85;
            break;
        case 'cyan':
            startHue = 128;
            break;
        case 'blue':
            startHue = 171;
            break;
        case 'magenta':
            startHue = 213;
            break;
        }
        originalNote = originalNote - Math.floor(originalNote); // strip off octave, use only note
        startHue = startHue + Math.floor(originalNote * 256);
        if (startHue > 255) {
            startHue = startHue - 256;
        }
        twist = Math.floor(instSpecificParams.twist * 128);
        stopHue = startHue + twist;
        if (stopHue > 255) {
            stopHue = stopHue - 256;
        }
        saturation = (1.0 - instSpecificParams.pastelness) * 254.0;
        saturation = Math.floor(saturation);
        bank = instSpecificParams.bank;
        unit = instSpecificParams.unit;
        umtSendLocalMsg(currentTime + "," + startMoment + "," + duration + ",danLights," + bank + "," + unit + "," + startHue + "," + saturation + "," + stopHue);
    };
}

function InstantiateFadeCandy() {
    "use strict";
    this.getParams = function () {
        return {
            percussion: true,
            fixed: [
                { name: "basecolor", display: "Base Color", type: "list", values: ["red=Red", "yellow=Yellow", "green=Green", "cyan=Cyan", "blue=Blue", "magenta=Magenta"], default: "red" },
                { name: "skip", display: "Skip", type: "list", values: ["1=1", "2=2", "3=3", "4=4", "5=5", "6=6", "7=7"], default: "2" },
                { name: "direction", display: "Direction", type: "list", values: ["-1=Up", "1=Down" ], default: "1" }
            ],
            parameters: [
                { name: "pastelness", display: "Pastel-ness" },
                { name: "twist", display: "Twist" }
            ]
        };
    };
    this.queUpANote = function (startMoment, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime) {
        var currentTime, startHue, stopHue, saturation, twist, skip, direction, remoteVoiceNum;
        console.log("startMoment");
        console.log(startMoment);
        console.log("frequency");
        console.log(frequency);
        console.log("duration");
        console.log(duration);
        console.log("amplitude");
        console.log(amplitude);
        console.log("instSpecificParams");
        console.log(instSpecificParams);
        console.log("originalNote");
        console.log(originalNote);
        console.log("originalStartTime");
        console.log(originalStartTime);
        console.log("-----");
        // time now, start time, duration, instrument, instrument specific params...
        // our inst specific params: hue, saturation
        currentTime = gUmt.globalCtx.currentTime;
        switch (instSpecificParams.basecolor) {
        case 'red':
            startHue = 0;
            remoteVoiceNum = 0;
            break;
        case 'yellow':
            startHue = 47;
            remoteVoiceNum = 1;
            break;
        case 'green':
            startHue = 85;
            remoteVoiceNum = 2;
            break;
        case 'cyan':
            startHue = 128;
            remoteVoiceNum = 3;
            break;
        case 'blue':
            startHue = 171;
            remoteVoiceNum = 4;
            break;
        case 'magenta':
            startHue = 213;
            remoteVoiceNum = 5;
            break;
        }
	remoteVoiceNum *= 6;
        skip = instSpecificParams.skip;
	remoteVoiceNum += skip;
        remoteVoiceNum *= 2;
        direction = instSpecificParams.direction;
	remoteVoiceNum += direction;
        originalNote = originalNote - Math.floor(originalNote); // strip off octave, use only note
        startHue = startHue + Math.floor(originalNote * 256);
        if (startHue > 255) {
            startHue = startHue - 256;
        }
        twist = Math.floor(instSpecificParams.twist * 128);
        stopHue = startHue + twist;
        if (stopHue > 255) {
            stopHue = stopHue - 256;
        }
        saturation = (1.0 - instSpecificParams.pastelness) * 254.0;
        saturation = Math.floor(saturation);
        umtSendLocalMsg(currentTime + "," + startMoment + "," + duration + ",fadeCandy," + remoteVoiceNum + "," + startHue + "," + saturation + "," + stopHue + "," + skip + "," + direction);
    };
}

function umtCreateInstrumentBank() {
    "use strict";
    var instrumentName;
    gUmt.instrumentBank = {
        tuningfork: new InstantiateTuningForkObj(),
        sinestartstop: new InstantiateSineStartStopObj(),
        squarewave: new InstantiateSquareWaveObj(),
        squarevar: new InstantiateSquareVarObj(),
        squarerisingpitch: new InstantiateSquareRisingPitchObj(),
        squarestartstop: new InstantiateSquareStartStopObj(),
        trianglewave: new InstantiateTriangleWaveObj(),
        trianglestartstop: new InstantiateTriangleStartStopObj(),
        sawtoothwave: new InstantiateSawtoothWaveObj(),
        sawtoothstartstop: new InstantiateSawtoothStartStopObj(),
        forwardsnoise: new InstantiateNoiseObj(),
        backwardsnoise: new InstantiateBackwardsNoiseObj(),
        steampop: new InstantiateSteamPopObj(),
        fmsynthoriginal: new InstantiateFMSynthOriginalObj(),
        trianglevar: new InstantiateTriangleVarObj(),
        sawtoothvar: new InstantiateSawtoothVarObj(),
        fmsynthvar: new InstantiateFMSynthVarObj(),
        fmsynthstartstop: new InstantiateFMSynthStartStopObj(),
        fmsynthsquare: new InstantiateFMSynthSquareObj(),
        fmsynthcrossnote: new InstantiateFMSynthCrossNoteObj(),
        wadsin: new InstantiateWadSinObj(),
        wadkick: new InstantiateWadKickObj(),
        wadbass: new InstantiateWadBassObj(),
        wadsnare: new InstantiateWadSnareObj(),
        wadhihatopen: new InstantiateWadHihatOpenObj(),
        wadhihatclosed: new InstantiateWadHihatClosedObj(),
        wadflute: new InstantiateWadFluteObj(),
        wadpiano: new InstantiateWadPianoObj(),
        danlights: new InstantiateDanLights(),
        fadecandy: new InstantiateFadeCandy()
    };
    gUmt.instrumentParams = {};
    for (instrumentName in gUmt.instrumentBank) {
        if (gUmt.instrumentBank.hasOwnProperty(instrumentName)) {
            gUmt.instrumentParams[instrumentName] = gUmt.instrumentBank[instrumentName].getParams();
        }
    }
}

function umtClearLoop(lpnum, tab, xTempo, centerNote, xScale, xChord) {
    "use strict";
    if (!gUmt.loop.hasOwnProperty(lpnum)) {
        gUmt.loop[lpnum] = {
            score: {
                playbackTimeOffset: 0,
                songTab: []
            }
        };
    }
    gUmt.loop[lpnum].score.songTab[tab] = {
        tempo: xTempo,
        centernoteLog: (Math.log(centerNote) / gUmt.LOG2) - 5,
        scale: xScale,
        chord: xChord,
        parts: 0,
        voice: []
    };
}

function umtAddVoice(lpnum, tab, idx, songNumber, rangeMin, rangeMax, instrumentName, xPercussion, xExemptFromHarmonization, xNoRepeatNotes, xRelativeNotes) {
    "use strict";
    gUmt.loop[lpnum].score.songTab[tab].parts = idx + 1;
    gUmt.loop[lpnum].score.songTab[tab].voice[idx] = {
        instrument: instrumentName,
        selection: 0,
        cursor: 0,
        nextStart: 0,
        rangeTop: rangeMax,
        rangeBottom: rangeMin,
        percussion: xPercussion,
        exemptFromHarmonization: xExemptFromHarmonization,
        noRepeatNotes: xNoRepeatNotes,
        relativeNotes: xRelativeNotes,
        notes: [],
        randRhythm: umtGetRando(songNumber * 40),
        randPitch: umtGetRando(songNumber * 40 + 1),
        randAmp: umtGetRando(songNumber * 40 + 5),
        randSkew: umtGetRando(songNumber * 40 + 13),
        randRests: umtGetRando(songNumber * 40 + 2),
        randSyms: umtGetRando(songNumber * 40 + 3),
        randTrans: umtGetRando(songNumber * 40 + 4),
        randScale: umtGetRando(songNumber * 40 + 12),
        randTilt: umtGetRando(songNumber * 40 + 6),
        randScrollrhythm: umtGetRando(songNumber * 40 + 7),
        randScrollpitch: umtGetRando(songNumber * 40 + 8),
        randScrollboth: umtGetRando(songNumber * 40 + 9),
        randInstPat: umtGetRando(songNumber * 40 + 10),
        randInstSymVars: umtGetRando(songNumber * 40 + 11)
    };
}

function umtAddNote(lpnum, tab, voicenum, rest, pitch, duration, amplitude, skew, instParams) {
    "use strict";
    var idx, cursorval;
    idx = voicenum;
    cursorval = gUmt.loop[lpnum].score.songTab[tab].voice[idx].cursor;
    gUmt.loop[lpnum].score.songTab[tab].voice[idx].notes[cursorval] = {
        starttime: gUmt.loop[lpnum].score.songTab[tab].voice[idx].nextStart,
        rest: rest,
        pitch: pitch,
        duration: duration,
        amplitude: amplitude,
        skew: skew,
        instParams: instParams,
        harmOctave: -1,
        harmNum: -1,
        harmDnom: -1
    };
    cursorval = cursorval + 1;
    gUmt.loop[lpnum].score.songTab[tab].voice[idx].cursor = cursorval;
    gUmt.loop[lpnum].score.songTab[tab].voice[idx].nextStart += duration;
}

// cheat sheet for scales
// ratio  interval  example key of C  logarithm  nearest equal temperment (logarithm)  difference
// 1/1    P1        C                 0.000      0.000      0.000
// 16/15  m2        C#                0.093      0.083      0.010
// 10/9   M2-alt    D-alt             0.152      0.166      0.015
// 9/8    M2-reg    D-reg             0.170      0.166      0.003
// 6/5    m3        Eb                0.263      0.250      0.013
// 5/4    M3        E                 0.322      0.333      0.011
// 4/3    P4        F                 0.415      0.417      0.002
// 45/32  A4-alt    F#                0.492      0.500      0.008
// 64/45  A4-reg    F#                0.508      0.500      0.008
// 3/2    P5        G                 0.585      0.583      0.002
// 8/5    m6        Ab                0.678      0.667      0.011
// 5/3    M6        A                 0.737      0.750      0.013
// 16/9   m7-reg    Bb-reg            0.830      0.833      0.003
// 9/5    m7-alt    Bb-alt            0.848      0.833      0.015
// 15/8   M7        B                 0.907      0.917      0.010
// 2/1    P8        C                 1.000      1.000      0.000

// Removed 25/18 (inverse 36/25) and 17/12 (inverse 24/17) as tritone and put in 64/45 (inverse 45/32)
// on the advice of David Genova.

function umtGetScale7Steps1() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 6, dnom: 5}, {num: 5, dnom: 4}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 8, dnom: 5}, {num: 5, dnom: 3} ];
    return scale;
}

function umtGetScale19Steps2() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 25, dnom: 24}, {num: 16, dnom: 15}, {num: 10, dnom: 9}, {num: 9, dnom: 8}, {num: 6, dnom: 5}, {num: 5, dnom: 4}, {num: 32, dnom: 25}, {num: 4, dnom: 3}, {num: 25, dnom: 18}, {num: 36, dnom: 25}, {num: 3, dnom: 2}, {num: 25, dnom: 16}, {num: 8, dnom: 5}, {num: 5, dnom: 3}, {num: 16, dnom: 9}, {num: 9, dnom: 5}, {num: 15, dnom: 8}, {num: 48, dnom: 25} ];
    return scale;
}

function umtGetScale37Steps3() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 128, dnom: 125}, {num: 25, dnom: 24}, {num: 16, dnom: 15}, {num: 27, dnom: 25}, {num: 10, dnom: 9}, {num: 9, dnom: 8}, {num: 144, dnom: 125}, {num: 125, dnom: 108}, {num: 75, dnom: 64}, {num: 32, dnom: 27}, {num: 6, dnom: 5}, {num: 5, dnom: 4}, {num: 32, dnom: 25}, {num: 125, dnom: 96}, {num: 4, dnom: 3}, {num: 27, dnom: 20}, {num: 25, dnom: 18}, {num: 45, dnom: 32}, {num: 64, dnom: 45}, {num: 36, dnom: 25}, {num: 40, dnom: 27}, {num: 3, dnom: 2}, {num: 192, dnom: 125}, {num: 25, dnom: 16}, {num: 8, dnom: 5}, {num: 5, dnom: 3}, {num: 27, dnom: 16}, {num: 128, dnom: 75}, {num: 216, dnom: 125}, {num: 125, dnom: 72}, {num: 16, dnom: 9}, {num: 9, dnom: 5}, {num: 50, dnom: 27}, {num: 15, dnom: 8}, {num: 48, dnom: 25}, {num: 125, dnom: 64} ];
    return scale;
}

function umtGetScale61Steps4() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 81, dnom: 80}, {num: 128, dnom: 125}, {num: 648, dnom: 625}, {num: 25, dnom: 24}, {num: 135, dnom: 128}, {num: 16, dnom: 15}, {num: 27, dnom: 25}, {num: 625, dnom: 576}, {num: 10, dnom: 9}, {num: 9, dnom: 8}, {num: 256, dnom: 225}, {num: 144, dnom: 125}, {num: 125, dnom: 108}, {num: 75, dnom: 64}, {num: 32, dnom: 27}, {num: 6, dnom: 5}, {num: 625, dnom: 512}, {num: 768, dnom: 625}, {num: 100, dnom: 81}, {num: 5, dnom: 4}, {num: 81, dnom: 64}, {num: 32, dnom: 25}, {num: 162, dnom: 125}, {num: 125, dnom: 96}, {num: 4, dnom: 3}, {num: 27, dnom: 20}, {num: 512, dnom: 375}, {num: 864, dnom: 625}, {num: 25, dnom: 18}, {num: 45, dnom: 32}, {num: 64, dnom: 45}, {num: 36, dnom: 25}, {num: 625, dnom: 432}, {num: 375, dnom: 256}, {num: 40, dnom: 27}, {num: 3, dnom: 2}, {num: 192, dnom: 125}, {num: 125, dnom: 81}, {num: 25, dnom: 16}, {num: 128, dnom: 81}, {num: 8, dnom: 5}, {num: 81, dnom: 50}, {num: 625, dnom: 384}, {num: 1024, dnom: 625}, {num: 5, dnom: 3}, {num: 27, dnom: 16}, {num: 128, dnom: 75}, {num: 216, dnom: 125}, {num: 125, dnom: 72}, {num: 225, dnom: 128}, {num: 16, dnom: 9}, {num: 9, dnom: 5}, {num: 1152, dnom: 625}, {num: 50, dnom: 27}, {num: 15, dnom: 8}, {num: 256, dnom: 135}, {num: 48, dnom: 25}, {num: 625, dnom: 324}, {num: 125, dnom: 64}, {num: 160, dnom: 81} ];
    return scale;
}

function umtGetScale91Steps5() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 81, dnom: 80}, {num: 3125, dnom: 3072}, {num: 128, dnom: 125}, {num: 250, dnom: 243}, {num: 648, dnom: 625}, {num: 25, dnom: 24}, {num: 256, dnom: 243}, {num: 135, dnom: 128}, {num: 16, dnom: 15}, {num: 27, dnom: 25}, {num: 625, dnom: 576}, {num: 2048, dnom: 1875}, {num: 1125, dnom: 1024}, {num: 3456, dnom: 3125}, {num: 10, dnom: 9}, {num: 9, dnom: 8}, {num: 256, dnom: 225}, {num: 144, dnom: 125}, {num: 125, dnom: 108}, {num: 75, dnom: 64}, {num: 32, dnom: 27}, {num: 6, dnom: 5}, {num: 3125, dnom: 2592}, {num: 243, dnom: 200}, {num: 625, dnom: 512}, {num: 768, dnom: 625}, {num: 100, dnom: 81}, {num: 3888, dnom: 3125}, {num: 5, dnom: 4}, {num: 512, dnom: 405}, {num: 81, dnom: 64}, {num: 32, dnom: 25}, {num: 625, dnom: 486}, {num: 162, dnom: 125}, {num: 125, dnom: 96}, {num: 4096, dnom: 3125}, {num: 320, dnom: 243}, {num: 675, dnom: 512}, {num: 4, dnom: 3}, {num: 27, dnom: 20}, {num: 3125, dnom: 2304}, {num: 512, dnom: 375}, {num: 864, dnom: 625}, {num: 25, dnom: 18}, {num: 45, dnom: 32}, {num: 64, dnom: 45}, {num: 36, dnom: 25}, {num: 625, dnom: 432}, {num: 375, dnom: 256}, {num: 4608, dnom: 3125}, {num: 40, dnom: 27}, {num: 3, dnom: 2}, {num: 1024, dnom: 675}, {num: 243, dnom: 160}, {num: 3125, dnom: 2048}, {num: 192, dnom: 125}, {num: 125, dnom: 81}, {num: 972, dnom: 625}, {num: 25, dnom: 16}, {num: 128, dnom: 81}, {num: 405, dnom: 256}, {num: 8, dnom: 5}, {num: 3125, dnom: 1944}, {num: 81, dnom: 50}, {num: 625, dnom: 384}, {num: 1024, dnom: 625}, {num: 400, dnom: 243}, {num: 5184, dnom: 3125}, {num: 5, dnom: 3}, {num: 27, dnom: 16}, {num: 128, dnom: 75}, {num: 216, dnom: 125}, {num: 125, dnom: 72}, {num: 225, dnom: 128}, {num: 16, dnom: 9}, {num: 9, dnom: 5}, {num: 3125, dnom: 1728}, {num: 2048, dnom: 1125}, {num: 1875, dnom: 1024}, {num: 1152, dnom: 625}, {num: 50, dnom: 27}, {num: 15, dnom: 8}, {num: 256, dnom: 135}, {num: 243, dnom: 128}, {num: 48, dnom: 25}, {num: 625, dnom: 324}, {num: 243, dnom: 125}, {num: 125, dnom: 64}, {num: 6144, dnom: 3125}, {num: 160, dnom: 81} ];
    return scale;
}

function umtGetScale15ChromaticTT() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 16, dnom: 15}, {num: 10, dnom: 9}, {num: 9, dnom: 8}, {num: 6, dnom: 5}, {num: 5, dnom: 4}, {num: 4, dnom: 3}, {num: 45, dnom: 32}, {num: 64, dnom: 45}, {num: 3, dnom: 2}, {num: 8, dnom: 5}, {num: 5, dnom: 3}, {num: 16, dnom: 9}, {num: 9, dnom: 5}, {num: 15, dnom: 8} ];
    return scale;
}

function umtGetScale13Chromatic() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 16, dnom: 15}, {num: 10, dnom: 9}, {num: 9, dnom: 8}, {num: 6, dnom: 5}, {num: 5, dnom: 4}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 8, dnom: 5}, {num: 5, dnom: 3}, {num: 16, dnom: 9}, {num: 9, dnom: 5}, {num: 15, dnom: 8} ];
    return scale;
}

function umtGetScale12ChromaticRegRegTT() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 16, dnom: 15}, {num: 9, dnom: 8}, {num: 6, dnom: 5}, {num: 5, dnom: 4}, {num: 4, dnom: 3}, {num: 64, dnom: 45}, {num: 3, dnom: 2}, {num: 8, dnom: 5}, {num: 5, dnom: 3}, {num: 16, dnom: 9}, {num: 15, dnom: 8} ];
    return scale;
}

function umtGetScale11ChromaticRegReg() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 16, dnom: 15}, {num: 9, dnom: 8}, {num: 6, dnom: 5}, {num: 5, dnom: 4}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 8, dnom: 5}, {num: 5, dnom: 3}, {num: 16, dnom: 9}, {num: 15, dnom: 8} ];
    return scale;
}

function umtGetScale11ChromaticAltReg() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 16, dnom: 15}, {num: 10, dnom: 9}, {num: 6, dnom: 5}, {num: 5, dnom: 4}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 8, dnom: 5}, {num: 5, dnom: 3}, {num: 16, dnom: 9}, {num: 9, dnom: 5}, {num: 15, dnom: 8} ];
    return scale;
}

function umtGetScale11ChromaticRegAlt() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 16, dnom: 15}, {num: 9, dnom: 8}, {num: 6, dnom: 5}, {num: 5, dnom: 4}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 8, dnom: 5}, {num: 5, dnom: 3}, {num: 16, dnom: 9}, {num: 15, dnom: 8} ];
    return scale;
}

function umtGetScale11ChromaticAltAlt() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 16, dnom: 15}, {num: 10, dnom: 9}, {num: 6, dnom: 5}, {num: 5, dnom: 4}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 8, dnom: 5}, {num: 5, dnom: 3}, {num: 16, dnom: 9}, {num: 15, dnom: 8} ];
    return scale;
}

function umtGetScale7MajorReg() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 9, dnom: 8}, {num: 5, dnom: 4}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 5, dnom: 3}, {num: 15, dnom: 8} ];
    return scale;
}

function umtGetScale7MajorAlt() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 10, dnom: 9}, {num: 5, dnom: 4}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 5, dnom: 3}, {num: 15, dnom: 8} ];
    return scale;
}

function umtGetScale7MinorNaturalRegReg() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 9, dnom: 8}, {num: 6, dnom: 5}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 8, dnom: 5}, {num: 16, dnom: 9} ];
    return scale;
}

function umtGetScale7MinorNaturalAltReg() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 10, dnom: 9}, {num: 6, dnom: 5}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 8, dnom: 5}, {num: 16, dnom: 9} ];
    return scale;
}

function umtGetScale7MinorNaturalRegAlt() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 9, dnom: 8}, {num: 6, dnom: 5}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 8, dnom: 5}, {num: 16, dnom: 9} ];
    return scale;
}

function umtGetScale7MinorNaturalAltAlt() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 10, dnom: 9}, {num: 6, dnom: 5}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 8, dnom: 5}, {num: 16, dnom: 9} ];
    return scale;
}

function umtGetScale7MinorHarmonicReg() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 9, dnom: 8}, {num: 6, dnom: 5}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 8, dnom: 5}, {num: 15, dnom: 8} ];
    return scale;
}

function umtGetScale7MinorHarmonicAlt() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 10, dnom: 9}, {num: 6, dnom: 5}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 8, dnom: 5}, {num: 15, dnom: 8} ];
    return scale;
}

function umtGetScale7MinorMelodicReg() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 9, dnom: 8}, {num: 6, dnom: 5}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 5, dnom: 3}, {num: 15, dnom: 8} ];
    return scale;
}

function umtGetScale7MinorMelodicAlt() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 10, dnom: 9}, {num: 6, dnom: 5}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 5, dnom: 3}, {num: 15, dnom: 8} ];
    return scale;
}

function umtGetScale7ArabicishMajor() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 16, dnom: 15}, {num: 5, dnom: 4}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 5, dnom: 3}, {num: 15, dnom: 8} ];
    return scale;
}

function umtGetScale7ArabicishMinor() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 16, dnom: 15}, {num: 5, dnom: 4}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 8, dnom: 5}, {num: 15, dnom: 8} ];
    return scale;
}

function umtGetScale5Pentatonic23Reg() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 9, dnom: 8}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 5, dnom: 3} ];
    return scale;
}

function umtGetScale5Pentatonic23Alt() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 10, dnom: 9}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 5, dnom: 3} ];
    return scale;
}

function umtGetScale5Pentatonic131Reg() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 6, dnom: 5}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 16, dnom: 9} ];
    return scale;
}

function umtGetScale5Pentatonic32Reg() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 9, dnom: 8}, {num: 5, dnom: 4}, {num: 3, dnom: 2}, {num: 5, dnom: 3} ];
    return scale;
}

function umtGetScale5Pentatonic32Alt() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 10, dnom: 9}, {num: 5, dnom: 4}, {num: 3, dnom: 2}, {num: 5, dnom: 3} ];
    return scale;
}

function umtGetScale5Pentatonic221Reg() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 9, dnom: 8}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 16, dnom: 9} ];
    return scale;
}

function umtGetScale5Pentatonic122Reg() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 6, dnom: 5}, {num: 4, dnom: 3}, {num: 8, dnom: 5}, {num: 16, dnom: 9} ];
    return scale;
}

function umtGetScale4Quadtonic() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 5, dnom: 3} ];
    return scale;
}

function umtGetScale3Tritonic() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 4, dnom: 3}, {num: 3, dnom: 2} ];
    return scale;
}

function umtGetScale2TritoneTest() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 64, dnom: 45} ];
    return scale;
}

function umtGetScale7ModeIonian() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 9, dnom: 8}, {num: 5, dnom: 4}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 5, dnom: 3}, {num: 15, dnom: 8} ];
    return scale;
}

function umtGetScale7ModeDorian() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 9, dnom: 8}, {num: 6, dnom: 5}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 5, dnom: 3}, {num: 16, dnom: 9} ];
    return scale;
}

function umtGetScale7ModePhrygian() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 16, dnom: 15}, {num: 6, dnom: 5}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 8, dnom: 5}, {num: 16, dnom: 9} ];
    return scale;
}

function umtGetScale7ModeLydianTT() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 9, dnom: 8}, {num: 5, dnom: 4}, {num: 64, dnom: 45}, {num: 3, dnom: 2}, {num: 5, dnom: 3}, {num: 15, dnom: 8} ];
    return scale;
}

function umtGetScale7ModeMixolydian() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 9, dnom: 8}, {num: 5, dnom: 4}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 5, dnom: 3}, {num: 16, dnom: 9} ];
    return scale;
}

function umtGetScale7ModeAeolian() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 9, dnom: 8}, {num: 6, dnom: 5}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 8, dnom: 5}, {num: 16, dnom: 9} ];
    return scale;
}

function umtGetScale7ModeLocrianTT() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 16, dnom: 15}, {num: 6, dnom: 5}, {num: 4, dnom: 3}, {num: 64, dnom: 45}, {num: 8, dnom: 5}, {num: 16, dnom: 9} ];
    return scale;
}

function umtGetScale7HungarianRegRegTT() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 9, dnom: 8}, {num: 6, dnom: 5}, {num: 64, dnom: 45}, {num: 3, dnom: 2}, {num: 8, dnom: 5}, {num: 15, dnom: 8} ];
    return scale;
}

function umtGetScale7Arabic() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 16, dnom: 15}, {num: 5, dnom: 4}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 8, dnom: 5}, {num: 15, dnom: 8} ];
    return scale;
}

function umtGetScale7FreygishPhrygianDominant() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 16, dnom: 15}, {num: 5, dnom: 4}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 8, dnom: 5}, {num: 16, dnom: 9} ];
    return scale;
}

function umtGetScale7UkranianDorian() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 9, dnom: 8}, {num: 6, dnom: 5}, {num: 64, dnom: 45}, {num: 3, dnom: 2}, {num: 5, dnom: 3}, {num: 16, dnom: 9} ];
    return scale;
}

function umtAddFloatsToScale(originalScale) {
    "use strict";
    var num, dnom, x, newScale, i;
    newScale = [];
    for (i = 0; i < originalScale.length; i = i + 1) {
        num = originalScale[i].num;
        dnom = originalScale[i].dnom;
        x = num / dnom;
        x = Math.log(x) / gUmt.LOG2;
        newScale[i] = {num: num, dnom: dnom, fpt: x};
    }
    return newScale;
}

function umtGaussGCD(a, b) {
    "use strict";
    // Gauss's GCD algorithm
    var c, d, e, r;
    c = a;
    d = b;
    r = 1;
    while (r !== 0) {
        if (d > c) {
            e = c;
            c = d;
            d = e;
        }
        r = (c % d);
        if (r !== 0) {
            c = d;
            d = r;
        }
    }
    return d;
}

// we are using Euclid's GCD algorithm because it should be faster for small numbers. Gauss's should be faster for large numbers
function umtEuclidGCD(a, b) {
    "use strict";
    // Euclid's GCD algorithm
    while (a !== b) {
        if (a > b) {
            a = a - b;
        } else {
            b = b - a;
        }
    }
    return a;
}

function umtReduceFraction(num, dnom) {
    "use strict";
    var g, rv;
    g = umtEuclidGCD(num, dnom);
    if (g !== 1) {
        num = num / g;
        dnom = dnom / g;
    }
    rv = {num: num, dnom: dnom};
    return rv;
}

function umtIsConsonant(lowNum, lowDnom, highNum, highDnom, rule) {
    "use strict";
    var ratioNum, ratioDnom, red;
    if (rule === "any") {
        return true;
    }
    ratioNum = highNum * lowDnom;
    ratioDnom = highDnom * lowNum;
    red = umtReduceFraction(ratioNum, ratioDnom);
    if (red.dnom > 3) {
        if (rule === "perfect") {
            return false;
        }
        if (red.dnom > 5) {
            if (rule === "major") {
                return false;
            }
        }
    }
    if (red.dnom > 6) {
        if (rule === "reg") {
            if ((red.dnom !== 9) && (red.dnom !== 15) && (red.dnom !== 18) && (red.dnom !== 25)) {
                return false;
            }
        } else {
            if (rule === "alt") {
                if ((red.dnom !== 8) && (red.dnom !== 9) && (red.dnom !== 15) && (red.dnom !== 16) && (red.dnom !== 18) && (red.dnom !== 25) && (red.dnom !== 27) && (red.dnom !== 32) && (red.dnom !== 45)) {
                    return false;
                }
            } else {
                return false;
            }
        }
    }
    ratioNum = 2 * lowNum * highDnom;
    ratioDnom = lowDnom * highNum;
    red = umtReduceFraction(ratioNum, ratioDnom);
    if (red.dnom > 3) {
        if (rule === "perfect") {
            return false;
        }
        if (red.dnom > 5) {
            if (rule === "major") {
                return false;
            }
        }
    }
    if (red.dnom > 6) {
        if (rule === "reg") {
            if ((red.dnom !== 9) && (red.dnom !== 15) && (red.dnom !== 18) && (red.dnom !== 25)) {
                return false;
            }
        } else {
            if (rule === "alt") {
                if ((red.dnom !== 8) && (red.dnom !== 9) && (red.dnom !== 15) && (red.dnom !== 16) && (red.dnom !== 18) && (red.dnom !== 25) && (red.dnom !== 27) && (red.dnom !== 32) && (red.dnom !== 45)) {
                    return false;
                }
            } else {
                return false;
            }
        }
    }
    return true;
}

function umtFractSubtract(firstNum, firstDnom, secondNum, secondDnom) {
    "use strict";
    // this function does not reduce the result!
    var g, s, xFNum, xFDnom, xSNum, xSDnom, rv;
    g = umtEuclidGCD(firstDnom, secondDnom);
    s = secondDnom / g;
    xFNum = firstNum * s;
    xFDnom = firstDnom * s;
    s = firstDnom / g;
    xSNum = secondNum * s;
    xSDnom = secondDnom * s;
    if (xFDnom !== xSDnom) {
        jsabort('denominators are not the same');
    }
    rv = { num: xFNum - xSNum, dnom: xFDnom };
    return rv;
}

function umtIsConsOrderInd(xnum, xdnom, ynum, ydnom, rule) {
    "use strict";
    var diff;
    // is consonant? order independent
    diff = umtFractSubtract(xnum, xdnom, ynum, ydnom);
    if (diff.num < 0) {
        // y is bigger
        return umtIsConsonant(xnum, xdnom, ynum, ydnom, rule);
    }
    // x is bigger
    return umtIsConsonant(ynum, ydnom, xnum, xdnom, rule);
}

// need caching for this function -- recalculating dissonance every time is computationally expensive
function umtDetermineIfFits(num, dnom, listOfNotesAlreadyPlaying, rule) {
    "use strict";
    var lx, i, diff, cacheidxname;
    cacheidxname = "udif" + num + "," + dnom + "," + rule;
    lx = listOfNotesAlreadyPlaying.length;
    for (i = 0; i < lx; i = i + 1) {
        cacheidxname = cacheidxname + "," + listOfNotesAlreadyPlaying[i].num + "," + listOfNotesAlreadyPlaying[i].dnom;
    }
    if (gUmt.cachedFittings.hasOwnProperty(cacheidxname)) {
        return gUmt.cachedFittings[cacheidxname];
    }
    for (i = 0; i < lx; i = i + 1) {
        diff = umtFractSubtract(num, dnom, listOfNotesAlreadyPlaying[i].num, listOfNotesAlreadyPlaying[i].dnom);
        if (diff.num < 0) {
            // already playing is bigger
            if (umtIsConsonant(num, dnom, listOfNotesAlreadyPlaying[i].num, listOfNotesAlreadyPlaying[i].dnom, rule) === false) {
                gUmt.cachedFittings[cacheidxname] = false;
                return false;
            }
        } else {
            // num/dom is bigger
            if (umtIsConsonant(listOfNotesAlreadyPlaying[i].num, listOfNotesAlreadyPlaying[i].dnom, num, dnom, rule) === false) {
                gUmt.cachedFittings[cacheidxname] = false;
                return false;
            }
        }
    }
    gUmt.cachedFittings[cacheidxname] = true;
    return true;
}

// just for detecting exclusions
function umtOnNoteList(num, dnom, noteList) {
    "use strict";
    var lx, i;
    if (noteList.length === 0) {
        return false;
    }
    lx = noteList.length;
    for (i = 0; i < lx; i = i + 1) {
        if ((num === noteList[i].harmNum) && (dnom === noteList[i].harmDnom)) {
            return true;
        }
    }
    return false;
}

function umtFindMostHarmonicNote(pitch, scale, listOfNotesAlreadyPlaying, rule, exclusionNotes) {
    "use strict";
    var octave, minIdx, minDist, i, dist, rv, originalFullPitch, numOfNotes, topPos, topOctave, bottomPos, bottomOctave, cycCount, topDist, bottomDist, currentPos, currentOctave;
    originalFullPitch = pitch;
    octave = Math.floor(pitch);
    pitch = pitch - octave;
    minIdx = -1;
    minDist = 1.0 - pitch;
    numOfNotes = scale.length;
    for (i = 0; i < numOfNotes; i = i + 1) {
        dist = Math.abs(scale[i].fpt - pitch);
        if (dist < minDist) {
            minIdx = i;
            minDist = dist;
        }
    }
    if ((listOfNotesAlreadyPlaying.length === 0) && (exclusionNotes.length === 0)) {
        if (minIdx === -1) {
            octave = octave + 1;
            rv = {octave: octave, scalenote: 0, num: scale[0].num, dnom: scale[0].dnom, fpt: scale[0].fpt };
        } else {
            rv = {octave: octave, scalenote: minIdx, num: scale[minIdx].num, dnom: scale[minIdx].dnom, fpt: scale[minIdx].fpt };
        }
        return rv;
    }
    // there are other notes playing: so we can't simply return the best scale
    // match. so we have to find another note.

    // the idea here is to "fan out" -- take the "min" note found above, start
    // with the scale notes directly above and below, and test the closest one
    // to see if it harmonizes with the existing notes. if it doesn't, push it
    // up/down and go through the loop again. mathematically we should be
    // guaranteed to find a match, even if it's a P1 with an existing note.
    // Brrrrp! -- the exclusion list for preventing note repeats, parallel
    // 5th/8ths/etc, breaks this guarantee we will find a note!
    if (minIdx === -1) {
        minIdx = 0;
        octave = octave + 1;
    }
    topPos = minIdx;
    topOctave = octave;
    bottomPos = minIdx;
    bottomOctave = octave;
    cycCount = 0;
    while (true) {
        cycCount = cycCount + 1;
        if (cycCount === 100) {
            jsabort('crash because stuck in endless loop (1864)');
        }
        topDist = Math.abs((topOctave + scale[topPos].fpt) - originalFullPitch);
        bottomDist = Math.abs((bottomOctave + scale[bottomPos].fpt) - originalFullPitch);
        if (topDist < bottomDist) {
            currentPos = topPos;
            currentOctave = topOctave;
            topPos = topPos + 1;
            if (topPos === numOfNotes) {
                topOctave = topOctave + 1;
                topPos = 0;
            }
        } else {
            currentPos = bottomPos;
            currentOctave = bottomOctave;
            bottomPos = bottomPos - 1;
            if (bottomPos < 0) {
                bottomOctave = bottomOctave - 1;
                bottomPos = numOfNotes - 1;
            }
        }
        if (!umtOnNoteList(scale[currentPos].num, scale[currentPos].dnom, exclusionNotes)) {
            if (umtDetermineIfFits(scale[currentPos].num, scale[currentPos].dnom, listOfNotesAlreadyPlaying, rule)) {
                rv = {octave: currentOctave, scalenote: currentPos, num: scale[currentPos].num, dnom: scale[currentPos].dnom, fpt: scale[currentPos].fpt };
                return rv;
            }
        }
        if ((topOctave === (octave + 1)) && (bottomOctave === (octave - 1))) {
            if ((topPos >= minIdx) && (bottomPos <= minIdx)) {
                // we have gone out of range! throw away the exclusion list and start over!
                exclusionNotes = [];
                topPos = minIdx;
                topOctave = octave;
                bottomPos = minIdx;
                bottomOctave = octave;
                // jsabort("We hit an exclusion list throwaway test case.");
            }
        }
    }
}

function umtGetScaleByAbbrv(scaleAbbrv) {
    "use strict";
    switch (scaleAbbrv) {
    case 'chroma13':
        return umtGetScale13Chromatic();
    case 'chroma11regreg':
        return umtGetScale11ChromaticRegReg();
    case 'maj7reg':
        return umtGetScale7MajorReg();
    case 'minnat7regreg':
        return umtGetScale7MinorNaturalRegReg();
    case 'minharm7reg':
        return umtGetScale7MinorHarmonicReg(); // use this one
    case 'minmelo7reg':
        return umtGetScale7MinorMelodicReg();
    case 'arabicmaj7':
        return umtGetScale7ArabicishMajor();
    case 'arabicmin7':
        return umtGetScale7ArabicishMinor();
    case 'pentatonic23reg':
        return umtGetScale5Pentatonic23Reg();
    case 'pentatonic131reg':
        return umtGetScale5Pentatonic131Reg();
    case 'pentatonic32reg':
        return umtGetScale5Pentatonic32Reg();
    case 'pentatonic221reg':
        return umtGetScale5Pentatonic221Reg();
    case 'pentatonic122reg':
        return umtGetScale5Pentatonic122Reg();
    case 'quad':
        return umtGetScale4Quadtonic();
    case 'tri':
        return umtGetScale3Tritonic();
    case 'chroma11altreg':
        return umtGetScale11ChromaticAltReg();
    case 'chroma11regalt':
        return umtGetScale11ChromaticRegAlt();
    case 'chroma11altalt':
        return umtGetScale11ChromaticAltAlt();
    case 'maj7alt':
        return umtGetScale7MajorAlt();
    case 'minnat7altreg':
        return umtGetScale7MinorNaturalAltReg();
    case 'minnat7regalt':
        return umtGetScale7MinorNaturalRegAlt();
    case 'minnat7altalt':
        return umtGetScale7MinorNaturalAltAlt();
    case 'minharm7alt':
        return umtGetScale7MinorHarmonicAlt();
    case 'minmelo7alt':
        return umtGetScale7MinorMelodicAlt();
    case 'pentatonic23alt':
        return umtGetScale5Pentatonic23Alt();
    case 'pentatonic32alt':
        return umtGetScale5Pentatonic32Alt();
    case 'tritone2test':
        return umtGetScale2TritoneTest();
    case 'ionian7':
        return umtGetScale7ModeIonian();
    case 'dorian7':
        return umtGetScale7ModeDorian();
    case 'phrygian7':
        return umtGetScale7ModePhrygian();
    case 'lydian7tt':
        return umtGetScale7ModeLydianTT();
    case 'mixolydian7':
        return umtGetScale7ModeMixolydian();
    case 'aeolian7':
        return umtGetScale7ModeAeolian();
    case 'locrian7tt':
        return umtGetScale7ModeLocrianTT();
    case 'hungarian7tt':
        return umtGetScale7HungarianRegRegTT();
    case 'chroma15tt':
        return umtGetScale15ChromaticTT();
    case 'chroma12regregtt':
        return umtGetScale12ChromaticRegRegTT();
    case 'arabic7':
        return umtGetScale7Arabic();
    case 'freygish7':
        return umtGetScale7FreygishPhrygianDominant();
    case 'ukranian7':
        return umtGetScale7UkranianDorian();
    case 'steps17':
        return umtGetScale7Steps1();
    case 'steps219':
        return umtGetScale19Steps2();
    case 'steps337':
        return umtGetScale37Steps3();
    case 'steps461':
        return umtGetScale61Steps4();
    case 'steps591':
        return umtGetScale91Steps5();
    default:
        jsabort("unrecognized scale code: " + scaleAbbrv);
        break;
    }
}

function umtHarmonizeScore(lpnum, tab) {
    "use strict";
    var theScale, rule, parts, vcnum, notecount, i, currentTime, currentPosition, timeRemaining, offEnd, stillGoing, maxFound, cycleCount, stillUnfoundMaxes, cycleUnfoundCount, maxIdx, poz, listOfNotesAlreadyPlaying, exclusionNotes, prevNotePosition, harmonizedInfo, minIdx, dist, lonapCursr;
    theScale = umtAddFloatsToScale(umtGetScaleByAbbrv(gUmt.loop[lpnum].score.songTab[tab].scale));
    rule = gUmt.loop[lpnum].score.songTab[tab].chord;
    // clear old data
    parts = gUmt.loop[lpnum].score.songTab[tab].parts;
    for (vcnum = 0; vcnum < parts; vcnum = vcnum + 1) {
        notecount = gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].cursor;
        for (i = 0; i < notecount; i = i + 1) {
            gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[i].harmOctave = -1;
            gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[i].harmNum = -1;
            gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[i].harmDnom = -1;
        }
    }
    // and off we go
    // this is a tricky function that absolutely relies on the position in the
    // notes array and the duration information being exactly right

    // the way it works is by setting current position to 0, and then advancing
    // the current position when time runs out, and setting the time remaining
    // to the duration of that note.

    // on each cycle, it subtracts the duration of the shortest note from each
    // of the 'time remaining' values for every note

    // however, while it uses the shortest note (minIdx) for subtracting time
    // (and thereby virtually 'plays' the piece), it uses the longest note
    // (maxIdx) for setting the harmony first

    // maxIdx is actually the longest "unharmonized" note -- not necessarily
    // the longest note
    currentTime = 0;
    currentPosition = [];
    for (i = 0; i < parts; i = i + 1) {
        currentPosition[i] = 0;
    }
    timeRemaining = [];
    for (i = 0; i < parts; i = i + 1) {
        timeRemaining[i] = gUmt.loop[lpnum].score.songTab[tab].voice[i].notes[currentPosition[i]].duration;
    }
    offEnd = [];
    for (i = 0; i < parts; i = i + 1) {
        if (currentPosition[i] === gUmt.loop[lpnum].score.songTab[tab].voice[i].cursor) {
            offEnd[i] = true;
        } else {
            offEnd[i] = false;
        }
    }
    stillGoing = false;
    for (i = 0; i < parts; i = i + 1) {
        if (offEnd[i] === false) {
            stillGoing = true;
        }
    }
    maxFound = [];
    cycleCount = 0;
    while (stillGoing) {
        cycleCount = cycleCount + 1;
        if (cycleCount > 3072) {
            jsabort('caught in endless loop?  cycle count blew up. (2112)');
            return;
        }
        for (i = 0; i < parts; i = i + 1) {
            maxFound[i] = false;
        }
        stillUnfoundMaxes = false;
        for (i = 0; i < parts; i = i + 1) {
            if (offEnd[i] === false) {
                if (maxFound[i] === false) {
                    stillUnfoundMaxes = true;
                }
            }
        }
        cycleUnfoundCount = 0;
        while (stillUnfoundMaxes) {
            cycleUnfoundCount = cycleUnfoundCount + 1;
            if (cycleUnfoundCount > 100) {
                jsabort('kot in endless looooop?? cycle count unfound blew up.');
                return;
            }
            // find max
            maxIdx = -1;
            for (i = 0; i < parts; i = i + 1) {
                if (offEnd[i] === false) {
                    if (maxFound[i] === false) {
                        if (maxIdx === -1) {
                            maxIdx = i;
                        } else {
                            if (timeRemaining[i] > timeRemaining[maxIdx]) {
                                maxIdx = i;
                            }
                        }
                    }
                }
            }
            if (maxIdx === -1) {
                jsabort('maxIdx is -1 -- should be impossible');
                return;
            }
            maxFound[maxIdx] = true;
            // here is the actual "harmonization" code -- everything else is
            // just looping
            if (gUmt.loop[lpnum].score.songTab[tab].voice[maxIdx].notes[currentPosition[maxIdx]].harmDnom === -1) {
                poz = currentPosition[maxIdx];
                // find all other notes that are turned on at the same time
                listOfNotesAlreadyPlaying = [];
                exclusionNotes = [];
                if (gUmt.loop[lpnum].score.songTab[tab].voice[maxIdx].noRepeatNotes) {
                    prevNotePosition = currentPosition[maxIdx] - 1;
                    if (prevNotePosition >= 0) {
                        if (gUmt.loop[lpnum].score.songTab[tab].voice[maxIdx].notes[prevNotePosition].rest === false) {
                            exclusionNotes = [ { harmNum: gUmt.loop[lpnum].score.songTab[tab].voice[maxIdx].notes[prevNotePosition].harmNum, harmDnom: gUmt.loop[lpnum].score.songTab[tab].voice[maxIdx].notes[currentPosition[maxIdx] - 1].harmDnom} ];
                        }
                    }
                }
                if ((gUmt.loop[lpnum].score.songTab[tab].voice[maxIdx].percussion) || (gUmt.loop[lpnum].score.songTab[tab].voice[maxIdx].exemptFromHarmonization)) {
                    // Brrrrrp -- if percussion or a voice exempt from harmonization, we skip the step of figuring out what other notes are playing, because it doesn't matter!
                    harmonizedInfo = umtFindMostHarmonicNote(gUmt.loop[lpnum].score.songTab[tab].voice[maxIdx].notes[poz].pitch, theScale, listOfNotesAlreadyPlaying, rule, exclusionNotes);
                } else {
                    lonapCursr = 0;
                    for (i = 0; i < parts; i = i + 1) {
                        if (currentPosition[i] < gUmt.loop[lpnum].score.songTab[tab].voice[i].cursor) { // check that we're not off the end
                            if (gUmt.loop[lpnum].score.songTab[tab].voice[i].notes[currentPosition[i]].harmDnom !== -1) {
                                // sanity check 2
                                if (currentTime < gUmt.loop[lpnum].score.songTab[tab].voice[i].notes[currentPosition[i]].starttime) {
                                    jsabort('START TIME IS AFTER CURRENT TIME! OH NO!');
                                }
                                if (currentTime >= (gUmt.loop[lpnum].score.songTab[tab].voice[i].notes[currentPosition[i]].starttime + gUmt.loop[lpnum].score.songTab[tab].voice[i].notes[currentPosition[i]].duration)) {
                                    jsabort('START TIME + DURATION IS BEFORE CURRENT TIME!! OH NO!!');
                                }
                                if (gUmt.loop[lpnum].score.songTab[tab].voice[i].percussion === false) {
                                    if (gUmt.loop[lpnum].score.songTab[tab].voice[i].exemptFromHarmonization === false) {
                                        if (gUmt.loop[lpnum].score.songTab[tab].voice[i].notes[currentPosition[i]].rest === false) {
                                            listOfNotesAlreadyPlaying[lonapCursr] = {
                                                octave: gUmt.loop[lpnum].score.songTab[tab].voice[i].notes[currentPosition[i]].harmOctave,
                                                num: gUmt.loop[lpnum].score.songTab[tab].voice[i].notes[currentPosition[i]].harmNum,
                                                dnom: gUmt.loop[lpnum].score.songTab[tab].voice[i].notes[currentPosition[i]].harmDnom
                                            };
                                            lonapCursr = lonapCursr + 1;
                                        }
                                    }
                                }
                            }
                        }
                    }
                    harmonizedInfo = umtFindMostHarmonicNote(gUmt.loop[lpnum].score.songTab[tab].voice[maxIdx].notes[poz].pitch, theScale, listOfNotesAlreadyPlaying, rule, exclusionNotes);
                }
                // sanity check 1
                if (currentTime !== gUmt.loop[lpnum].score.songTab[tab].voice[maxIdx].notes[currentPosition[maxIdx]].starttime) {
                    cx('START TIMES DO NOT MATCH! OH NO!');
                    cx('currentTime is ' + currentTime);
                    cx('gUmt.loop[lpnum].score.songTab[tab].voice[maxIdx].notes[currentPosition[maxIdx]].starttime is ' + gUmt.loop[lpnum].score.songTab[tab].voice[maxIdx].notes[currentPosition[maxIdx]].starttime);
                    cx('lpnum is ' + lpnum);
                    cx('maxIdx is ' + maxIdx);
                    cx('currentPosition[maxIdx] is ' + currentPosition[maxIdx]);
                    console.log('gUmt', gUmt);
                    jsabort('START TIMES DO NOT MATCH! OH NO!');
                }
                gUmt.loop[lpnum].score.songTab[tab].voice[maxIdx].notes[currentPosition[maxIdx]].harmOctave = harmonizedInfo.octave;
                gUmt.loop[lpnum].score.songTab[tab].voice[maxIdx].notes[currentPosition[maxIdx]].harmNum = harmonizedInfo.num;
                gUmt.loop[lpnum].score.songTab[tab].voice[maxIdx].notes[currentPosition[maxIdx]].harmDnom = harmonizedInfo.dnom;
                // umtSanitTestEntireScore(lpnum, false);
            }
            // end of actual "harmonization" code"
            stillUnfoundMaxes = false; // this loop could no doubt be optimized
            for (i = 0; i < parts; i = i + 1) {
                if (offEnd[i] === false) {
                    if (maxFound[i] === false) {
                        stillUnfoundMaxes = true;
                    }
                }
            }
        }
        // find min
        minIdx = -1;
        for (i = 0; i < parts; i = i + 1) {
            if (offEnd[i] === false) {
                if (minIdx === -1) {
                    minIdx = i;
                } else {
                    if (timeRemaining[i] < timeRemaining[minIdx]) {
                        minIdx = i;
                    }
                }
            }
        }
        // subtract dist
        dist = timeRemaining[minIdx];
        currentTime += dist;
        for (i = 0; i < parts; i = i + 1) {
            if (offEnd[i] === false) {
                timeRemaining[i] -= dist;
            }
        }
        // advance positions
        for (i = 0; i < parts; i = i + 1) {
            if (offEnd[i] === false) {
                if (timeRemaining[i] === 0) {
                    currentPosition[i] = currentPosition[i] + 1;
                    if (currentPosition[i] === gUmt.loop[lpnum].score.songTab[tab].voice[i].cursor) {
                        offEnd[i] = true;
                    } else {
                        timeRemaining[i] = gUmt.loop[lpnum].score.songTab[tab].voice[i].notes[currentPosition[i]].duration;
                    }
                }
            }
        }
        for (i = 0; i < parts; i = i + 1) {
            if (currentPosition[i] === gUmt.loop[lpnum].score.songTab[tab].voice[i].cursor) {
                offEnd[i] = true;
            }
        }
        stillGoing = false;
        for (i = 0; i < parts; i = i + 1) {
            if (offEnd[i] === false) {
                stillGoing = true;
            }
        }
    }
    // umtSanitTestEntireScore(lpnum, true);
    cx('umtHarmonizeScore(lpnum) done');
}

function umtExecHarmonizeScore() {
    "use strict";
    umtHarmonizeScore(0, 0);
}

function umtGenRhythmSubset(randomObj, minSize, maxSize, chunkSize) {
    "use strict";
    var setSize, setAndMax, rv, currentTime, position, amountLeft, dur;
    setSize = chunkSize / minSize;
    setAndMax = maxSize / minSize;
    rv = [];
    currentTime = 0;
    position = 0;
    while (currentTime < setSize) {
        amountLeft = setSize - currentTime;
        if (amountLeft === 1) {
            dur = 1;
        } else {
            if (amountLeft > setAndMax) {
                dur = Math.floor(randomObj.genrandReal2() * setAndMax + 1);
            } else {
                dur = Math.floor(randomObj.genrandReal2() * amountLeft + 1);
            }
        }
        rv[position] = dur * minSize;
        position = position + 1;
        currentTime = currentTime + dur;
    }
    return rv;
}

function umtGenerateRhythm(randomObj, minSize, maxSize, chunkSize, totalLength) {
    "use strict";
    var rv, currentStart, chunkData, endPt, chunkLen, idx;
    if (chunkSize === totalLength) {
        return umtGenRhythmSubset(randomObj, minSize, maxSize, chunkSize);
    }
    rv = [];
    for (currentStart = 0; currentStart < totalLength; currentStart = currentStart + chunkSize) {
        chunkData = umtGenRhythmSubset(randomObj, minSize, maxSize, chunkSize);
        endPt = rv.length;
        chunkLen = chunkData.length;
        for (idx = 0; idx < chunkLen; idx = idx + 1) {
            rv[endPt + idx] = chunkData[idx];
        }
    }
    return rv;
}

function umtAddSymmetryBaseToVoice(lpnum, tab, vcnum, minSize, maxSize, chunkSize, totalLength, centerOctave, restyness, noteDistance, relativeNotes, ampVariation, volume, voiceSkew, instrumentName, instrumentParamNames, instSpecificParams, instFixedParams) {
    "use strict";
    var rhythm, lenRhy, idx, duration, pitch, amplitude, noteSkew, volscalar, instParamCurrent, instParCount, instParamIdx, ourParamName, paramvalue, instFixedIdx, ourFixedValue, halfVoiceSkew, previousPitch;
    volscalar = 2.7;
    volume = Math.exp(volume * volscalar) / Math.exp(volscalar); // this is done to "logarithmize" the volume slider
    rhythm = umtGenerateRhythm(gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].randRhythm, minSize, maxSize, chunkSize, totalLength);
    lenRhy = rhythm.length;
    halfVoiceSkew = voiceSkew / 2.0;
    previousPitch = 0.0;
    for (idx = 0; idx < lenRhy; idx = idx + 1) {
        duration = rhythm[idx];
        if (relativeNotes) {
            pitch = (gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].randPitch.genrandReal2() * (2 * noteDistance)) - noteDistance + (previousPitch * (1.0 - noteDistance));
        } else {
            pitch = (gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].randPitch.genrandReal2() * (2 * noteDistance)) - noteDistance;
        }
        if (gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].randRests.genrandReal2() < restyness) {
            umtAddNote(lpnum, tab, vcnum, true, centerOctave, duration, 1, 0, null);
        } else {
            amplitude = (1.0 - (gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].randAmp.genrandReal2() * ampVariation) * 0.9) * volume;
            noteSkew = gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].randSkew.genrandReal2() * voiceSkew;
            if (noteSkew > halfVoiceSkew) {
                // This is so we can get both positive and negative numbers, symmetric around 0, from a single call to genrandReal2()
                noteSkew = noteSkew - voiceSkew;
            }
            // The idea here is to go through the param names for this instrument and pull out the corresponding "patterns" value from instSpecificParams
            // We ignore all parameters that don't apply to the current instrument. This way the user can flip between instruments and the system won't forget their previous settings.
            instParamCurrent = {};
            instParCount = 0;
            for (instParamIdx in instrumentParamNames) {
                if (instrumentParamNames.hasOwnProperty(instParamIdx)) {
                    ourParamName = instrumentParamNames[instParamIdx].name;
                    if (instSpecificParams.hasOwnProperty(instrumentName + "_" + ourParamName + "_patterns")) {
                        paramvalue = instSpecificParams[instrumentName + "_" + ourParamName + "_patterns"];
                        instParamCurrent[ourParamName] = gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].randInstPat.genrandReal2() * paramvalue;
                        instParCount = instParCount + 1;
                    }
                }
            }
            // we just copy in the fixed parameters here -- it wastes space in the composition and we should probably put this at a per-voice level, but we're lazy, so we don't
            for (instFixedIdx in instFixedParams) {
                if (instFixedParams.hasOwnProperty(instFixedIdx)) {
                    if (instFixedIdx.substring(0, instrumentName.length) === instrumentName) {
                        ourFixedValue = instFixedParams[instFixedIdx];
                        instParamCurrent[instFixedIdx.substring(instrumentName.length + 1)] = ourFixedValue;
                        instParCount = instParCount + 1;
                    }
                }
            }
            if (instParCount === 0) {
                instParamCurrent = null;
            }
            umtAddNote(lpnum, tab, vcnum, false, centerOctave + pitch, duration, amplitude, noteSkew, instParamCurrent);
            previousPitch = pitch;
        }
    }
}

function umtSchedulePlayOfNote(lpnum, tab, voicenum, notenum, ampScalar, tempoScalar, center, adjustedStartTime, originalStartTime) {
    "use strict";
    var harmonizedOctave, harmonizedNum, harmonizedDnom, duration, amplitude, instSpecificParams, frequency, instrument, originalNote;
    if (gUmt.loop[lpnum].score.songTab[tab].voice[voicenum].notes[notenum].rest === false) {
        harmonizedOctave = gUmt.loop[lpnum].score.songTab[tab].voice[voicenum].notes[notenum].harmOctave;
        harmonizedNum = gUmt.loop[lpnum].score.songTab[tab].voice[voicenum].notes[notenum].harmNum;
        harmonizedDnom = gUmt.loop[lpnum].score.songTab[tab].voice[voicenum].notes[notenum].harmDnom;
        duration = gUmt.loop[lpnum].score.songTab[tab].voice[voicenum].notes[notenum].duration * tempoScalar;
        amplitude = gUmt.loop[lpnum].score.songTab[tab].voice[voicenum].notes[notenum].amplitude;
        amplitude = (amplitude * Math.exp((amplitude - 1.0) * 2.0)) * ampScalar;
        instSpecificParams = gUmt.loop[lpnum].score.songTab[tab].voice[voicenum].notes[notenum].instParams;
        frequency = Math.exp((harmonizedOctave + center) * gUmt.LOG2);
        frequency = frequency * (harmonizedNum / harmonizedDnom);
        instrument = gUmt.loop[lpnum].score.songTab[tab].voice[voicenum].instrument;
        // umtTestDetectDuplicate(instrument, adjustedStartTime, frequency, duration, amplitude);
        originalNote = gUmt.loop[lpnum].score.songTab[tab].voice[voicenum].notes[notenum].pitch;
        gUmt.instrumentBank[instrument].queUpANote(adjustedStartTime, frequency, duration, amplitude, instSpecificParams, originalNote, originalStartTime);
    }
}

function umtCalculateMultiTabLoopLength(lpnum) {
    "use strict";
    var total, tab, numTabs, nxtTab, tabUnitLen, tempoScalarThisTab, tempoScalarNextTab, k1;
    numTabs = gUmt.loop[lpnum].score.numTabs;
    total = 0;
    gUmt.loop[lpnum].score.tabStartTimeOffsets = [];
    for (tab = 0; tab < numTabs; tab = tab + 1) {
        // in addition to calculating the total, we store where we are at each
        // tab as we go along, so that we don't have to recalculate it later for
        // every single note
        gUmt.loop[lpnum].score.tabStartTimeOffsets[tab] = total;
        tabUnitLen = gUmt.loop[lpnum].score.songTab[tab].voice[0].nextStart; // a bit of a hack, grabbing the start position of the next note, were we to compose one
        tempoScalarThisTab = 60 / gUmt.loop[lpnum].score.songTab[tab].tempo;
        nxtTab = tab + 1;
        if (nxtTab === numTabs) {
            nxtTab = 0;
        }
        tempoScalarNextTab = 60 / gUmt.loop[lpnum].score.songTab[nxtTab].tempo;
        if (tempoScalarThisTab === tempoScalarNextTab) {
            total = total + (tempoScalarThisTab * tabUnitLen);
        } else {
            // a bit of calculus to handle gradual tempo change
            k1 = (tempoScalarNextTab - tempoScalarThisTab) / tabUnitLen;
            total = total + ((k1 * tabUnitLen * tabUnitLen) / 2.0) + (tempoScalarThisTab * tabUnitLen);
        }
    }
    return total;
}

function umtPlaybackFigureOutWhichTabWeAreOn(currentTime, numTabs, lpnum, loopLen) {
    "use strict";
    var floatLoopTabCycles, numLoopTabCycles, positionInLoop, tab, rv;
    floatLoopTabCycles = currentTime / loopLen;
    numLoopTabCycles = Math.floor(floatLoopTabCycles);
    positionInLoop = currentTime - (numLoopTabCycles * loopLen);
    rv = 0;
    for (tab = 0; tab < numTabs; tab = tab + 1) {
        if (gUmt.loop[lpnum].score.tabStartTimeOffsets[tab] <= positionInLoop) {
            rv = tab;
        }
    }
    return rv;
}

function umtPlaybackHighlightOneTab(tabToHighlight) {
    "use strict";
    var start, children, i, j;
    start = jQuery(".ui-tabs-nav").get(0); // start is a <ul> tag
    if (start.hasChildNodes()) {
        // should always be true
        children = start.childNodes; // children are <li> tags
        j = 0; // count non-text nodes
        for (i = 0; i < children.length; i = i + 1) {
            if (children[i].nodeType === 1) {
                if (j === tabToHighlight) {
                    children[i].childNodes[0].style.backgroundColor = "#000000";
                } else {
                    children[i].childNodes[0].style.backgroundColor = "#FFFFFF";
                }
                j = j + 1;
            }
        }
    }
}

function umtPlaybackHighlightPlayingTab(currentTime, numTabs, lpnum, loopLen) {
    "use strict";
    var tabToHighlight;
    // Some rather hairy code here that depends on the internal structure of the Tab HTML generated by jQuery!
    // If you ever upgrade the jQuery / jQuery UI used for this project, this code may break!
    tabToHighlight = umtPlaybackFigureOutWhichTabWeAreOn(currentTime, numTabs, lpnum, loopLen);
    umtPlaybackHighlightOneTab(tabToHighlight);
}

var umtExecReSeed;

function umtSchedulePlayOfSegment() {
    "use strict";
    var tab, nxTab, lpnum, currentTime, compositonlag, windowStart, windowStop, loopUnitLen, loopLen, numLoopTabCycles, center, tempoScalar, tempoThisScalar, tempoNextScalar, k1, numVoices, voicenum, notenum, originalStartTime, adjustedStartTime, masterVolScaler, ampScalar, regularTime, timeInMs, interval, numTabs, floatLoopTabCycles, tabUnitLen;
    lpnum = gUmt.currentlyPlayingLoop;
    currentTime = gUmt.globalCtx.currentTime;
    // figure out our playback window for the segment we are going to play in this timer callback
    windowStart = currentTime;
    compositonlag = gUmt.compositionLag;
    if (compositonlag < 0.1) {
        compositonlag = 0.1;
    }
    if (windowStart < gUmt.playedUpTo) {
        windowStart = gUmt.playedUpTo;
    }
    windowStop = currentTime + (compositonlag * 2.5);
    if (gUmt.loop[lpnum].score.playAllTabs) {
        // cx("Scheduling playback of segment from tab ALL, loop " + ctstr(lpnum) + ", windowStart " + ctstr(windowStart) + ", windowStop " + ctstr(windowStop));
    } else {
        tab = gUmt.UIParams.currentTab;
        // cx("Scheduling playback of segment from tab " + ctstr(tab) + ", loop " + ctstr(lpnum) + ", windowStart " + ctstr(windowStart) + ", windowStop " + ctstr(windowStop));
    }
    if (windowStop < gUmt.playedUpTo) {
        cx("Nothing to play!");
    } else {
        if (windowStart > windowStop) {
            jsabort("Start and stop times got crossed!");
        }
        // stuff that applies to all notes
        masterVolScaler = gUmt.UIParams.mastervol; // Math.exp((gUmt.UIParams.mastervol - 1.0) * 2.0);
        // now we fork off "all tabs" vs "current tab"
        if (gUmt.loop[lpnum].score.playAllTabs) {
            numTabs = gUmt.loop[lpnum].score.numTabs;
            loopLen = umtCalculateMultiTabLoopLength(lpnum); // qx tabLen * numTabs;
            umtPlaybackHighlightPlayingTab(currentTime, numTabs, lpnum, loopLen);
            for (tab = 0; tab < numTabs; tab = tab + 1) {
                if (gUmt.loop[gUmt.currentlyPlayingLoop].score.songTab[tab].parts > 0) {
                    tabUnitLen = gUmt.loop[lpnum].score.songTab[tab].voice[0].nextStart; // a bit of a hack, grabbing the start position of the next note, were we to compose one
                    nxTab = tab + 1;
                    if (nxTab === numTabs) {
                        nxTab = 0;
                    }
                    tempoThisScalar = 60 / gUmt.loop[lpnum].score.songTab[tab].tempo;
                    tempoNextScalar = 60 / gUmt.loop[lpnum].score.songTab[nxTab].tempo;
                    // do a little calculus to get the tempo to transition smoothly from one tab to the next
                    k1 = ((tempoNextScalar - tempoThisScalar) / tabUnitLen) / 2.0;
                    // note time is returned by (k1 * tnote * tnote) + (tempoThisScalar * tnote);
                    center = gUmt.loop[lpnum].score.songTab[tab].centernoteLog;
                    numVoices = gUmt.loop[lpnum].score.songTab[tab].parts;
                    if (numVoices === 1) {
                        ampScalar = masterVolScaler / 1.4142135623730951;
                    } else {
                        ampScalar = masterVolScaler / numVoices;
                    }
                    for (voicenum = 0; voicenum < numVoices; voicenum = voicenum + 1) {
                        for (notenum = 0; notenum < gUmt.loop[lpnum].score.songTab[tab].voice[voicenum].cursor; notenum = notenum + 1) {
                            originalStartTime = gUmt.loop[lpnum].score.songTab[tab].voice[voicenum].notes[notenum].starttime;
                            // adjust for skew
                            adjustedStartTime = originalStartTime + gUmt.loop[lpnum].score.songTab[tab].voice[voicenum].notes[notenum].skew;
                            // adjust for tempo
                            tempoScalar = (adjustedStartTime * (k1 * 2)) + tempoThisScalar;
                            adjustedStartTime = (k1 * adjustedStartTime * adjustedStartTime) + (tempoThisScalar * adjustedStartTime);
                            // adjustedStartTime = adjustedStartTime + (tabLen * tab);
                            adjustedStartTime = gUmt.loop[lpnum].score.tabStartTimeOffsets[tab] + adjustedStartTime;
                            floatLoopTabCycles = (windowStart - adjustedStartTime) / loopLen;
                            numLoopTabCycles = Math.floor(floatLoopTabCycles);
                            if (numLoopTabCycles === undefined) {
                                jsabort("numLoopTapCycles is undefined");
                            }
                            if (floatLoopTabCycles !== numLoopTabCycles) {
                                numLoopTabCycles = numLoopTabCycles + 1;
                            }
                            if (numLoopTabCycles === undefined) {
                                jsabort("numLoopTapCycles is undefined");
                            }
                            adjustedStartTime = adjustedStartTime + (loopLen * numLoopTabCycles);
                            // play as many times as can fit in the current playback window
                            while (adjustedStartTime < windowStop) {
                                umtSchedulePlayOfNote(lpnum, tab, voicenum, notenum, ampScalar, tempoScalar, center, adjustedStartTime, originalStartTime);
                                adjustedStartTime = adjustedStartTime + loopLen;
                            }
                            if (numLoopTabCycles === undefined) {
                                jsabort("numLoopTapCycles is undefined");
                            }
                        }
                    }
                }
            }
        } else {
            tab = gUmt.UIParams.currentTab;
            tempoScalar = 60 / gUmt.loop[lpnum].score.songTab[tab].tempo;
            if (gUmt.loop[gUmt.currentlyPlayingLoop].score.songTab[tab].parts > 0) {
                numVoices = gUmt.loop[lpnum].score.songTab[tab].parts;
                loopUnitLen = gUmt.loop[lpnum].score.songTab[tab].voice[0].nextStart; // a bit of a hack, grabbing the start position of the next note, were we to compose one
                loopLen = loopUnitLen * tempoScalar;
                center = gUmt.loop[lpnum].score.songTab[tab].centernoteLog;
                if (numVoices === 1) {
                    ampScalar = masterVolScaler / 1.4142135623730951;
                } else {
                    ampScalar = masterVolScaler / numVoices;
                }
                for (voicenum = 0; voicenum < numVoices; voicenum = voicenum + 1) {
                    for (notenum = 0; notenum < gUmt.loop[lpnum].score.songTab[tab].voice[voicenum].cursor; notenum = notenum + 1) {
                        originalStartTime = gUmt.loop[lpnum].score.songTab[tab].voice[voicenum].notes[notenum].starttime;
                        // adjust for skew
                        adjustedStartTime = originalStartTime + gUmt.loop[lpnum].score.songTab[tab].voice[voicenum].notes[notenum].skew;
			// adjust to tempo
                        adjustedStartTime = adjustedStartTime * tempoScalar;
                        // loop is one tab, so move to place within loop
                        floatLoopTabCycles = (windowStart - adjustedStartTime) / loopLen;
                        numLoopTabCycles = Math.floor(floatLoopTabCycles);
                        if (floatLoopTabCycles !== numLoopTabCycles) {
                            numLoopTabCycles = numLoopTabCycles + 1;
                        }
                        adjustedStartTime = adjustedStartTime + (loopLen * numLoopTabCycles);
                        // schedule for playback just once (or not at all if note is already outside the window)
                        if (adjustedStartTime < windowStop) {
                            umtSchedulePlayOfNote(lpnum, tab, voicenum, notenum, ampScalar, tempoScalar, center, adjustedStartTime, originalStartTime);
                        }
                    }
                }
            }
        }
        // remember where we played up to, so we don't repeat notes
        gUmt.playedUpTo = windowStop;
    }
    // hack -- UI logic in the playback system?
    if (document.getElementById("continuous").checked) {
        regularTime = new Date(); // year, month [, day, hour, minute, second, millisecond]);
        timeInMs = regularTime.getTime();
        interval = timeInMs - gUmt.lastReSeedTime;
        cx(interval);
        if ((interval > 120000) || (interval < 0)) {
            umtExecReSeed();
        }
    }
    if (gUmt.timerIsOn) {
        if (compositonlag > 1.0) {
            compositonlag = 1.0; // force this timer function to be called within a certain frequency in attempt to plug gaps in the playback
        }
        window.setTimeout(umtSchedulePlayOfSegment, compositonlag * 1000);
    }
}

function umtDoTimerPulse() {
    "use strict";
    umtSchedulePlayOfSegment();
}

function umtAutomaticStart(lpnum) {
    "use strict";
    gUmt.currentlyPlayingLoop = lpnum;
    if (!gUmt.timerIsOn) {
        window.setTimeout(umtDoTimerPulse, 500);
    }
    gUmt.timerIsOn = true;
}

function umtAutomaticStop() {
    "use strict";
    gUmt.timerIsOn = false;
    gUmt.compositionLag = 0.0;
}

function umtCopyAndTransform(lpnum, tab, vcnum, startTime, endTime, rhythmDirection, pitchesDirection, timeRhythmScroll, timePitchScroll, actualPitchScroll, pitchScale, pitchBevel, pitchTranspose) {
    "use strict";
    var segmentStart, ix, segmentEnd, tsRhCursor, tsPiCursor, minPitch, maxPitch, thisPitch, midpointPitch, rhythmCursor, pitchCursor, terminationCursor, rest, duration, amplitude, skew, instParams, percent, bevelAmount, segmentLen, rangeTestStart, rangeTestEnd, rangeTestCursor, rangePitch, rangeMin, rangeMax, rangeScalar, rangeTranspose;
    // first find the start and end of the segment, and the starting point for the rhythm cursor (tsRhCursor) and pitch cursor (tsPiCursor) which can be different if there is rhythm or pitch scrolling
    segmentStart = -1;
    tsRhCursor = -1; // ts == "time scroll"; rh == "rhythm", pi == "pitch"
    tsPiCursor = -1;
    for (ix = 0; ix < gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].cursor; ix = ix + 1) {
        // yes, we search the whole voice -- this is ok because loops are short.
        // If the sytem is ever used for long compositions, this will need to
        // be replaced with interval halving
        if (segmentStart === -1) {
            if (gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[ix].starttime >= startTime) {
                segmentStart = ix;
                maxPitch = gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[ix].pitch;
                minPitch = maxPitch;
            }
        }
        if (gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[ix].starttime < endTime) {
            segmentEnd = ix;
        }
        if (tsRhCursor === -1) {
            if (gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[ix].starttime >= (startTime + timeRhythmScroll)) {
                tsRhCursor = ix;
            }
        }
        if (tsPiCursor === -1) {
            if (gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[ix].starttime >= (startTime + timePitchScroll)) {
                tsPiCursor = ix;
            }
        }
        if ((gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[ix].starttime >= startTime) && (gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[ix].starttime < endTime)) {
            // if notes are "playing" get min and max
            // we need min and max for scrolling effect
            thisPitch = gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[ix].pitch;
            if (thisPitch > maxPitch) {
                maxPitch = thisPitch;
            }
            if (thisPitch < minPitch) {
                minPitch = thisPitch;
            }
        }
    }
    if (tsRhCursor === -1) {
        tsRhCursor = 0;
    }
    if (tsPiCursor === -1) {
        tsPiCursor = 0;
    }
    segmentLen = segmentEnd - segmentStart + 1;
    midpointPitch = (minPitch + maxPitch) / 2;
    // time scroll cursors handle "horizontal" scrolling (pitch and rhythm as separate channels)
    rhythmCursor = tsRhCursor;
    pitchCursor = tsPiCursor;
    if (rhythmDirection === -1) {
        rhythmCursor = segmentEnd - (tsRhCursor - segmentStart);
    }
    if (pitchesDirection === -1) {
        pitchCursor = segmentEnd - (tsPiCursor - segmentStart);
    }
    terminationCursor = segmentStart;
    rangeTestStart = gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].cursor;
    // execute the copy+transform
    while (terminationCursor <= segmentEnd) {
        rest = gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[pitchCursor].rest;
        thisPitch = gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[pitchCursor].pitch;
        duration = gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[rhythmCursor].duration;
        amplitude = gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[rhythmCursor].amplitude;
        skew = gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[rhythmCursor].skew;
        instParams = gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[rhythmCursor].instParams;
        // actual pitch scroll handles "vertical" scrolling
        if (actualPitchScroll !== 0) {
            thisPitch = thisPitch    + actualPitchScroll;
            if (thisPitch > maxPitch) {
                thisPitch = thisPitch - (maxPitch - minPitch);
            } else {
                if (thisPitch < minPitch) {
                    thisPitch = thisPitch + (maxPitch - minPitch);
                }
            }
        }
        if (pitchScale !== 1) {
            thisPitch = ((thisPitch - midpointPitch) * pitchScale) + midpointPitch;
        }
        // "bevel" bevels only towards the end note, but bevel combined with
        // transpose can bevel both ends; having an additional parameter to
        // bevel the front end would be redundant because it would just be
        // subtracted from transpose
        if (pitchBevel !== 0) {
            percent = (gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[terminationCursor].starttime - startTime) / (endTime - startTime);
            bevelAmount = pitchBevel * percent;
            thisPitch = thisPitch + bevelAmount;
        }
        if (pitchTranspose !== 0) {
            thisPitch = thisPitch + pitchTranspose;
        }
        umtAddNote(lpnum, tab, vcnum, rest, thisPitch, duration, amplitude, skew, instParams);
        rhythmCursor = rhythmCursor + rhythmDirection;
        if (rhythmCursor > segmentEnd) {
            rhythmCursor = rhythmCursor - segmentLen;
        }
        if (rhythmCursor < segmentStart) {
            rhythmCursor = pitchCursor + segmentLen;
        }
        pitchCursor = pitchCursor + pitchesDirection;
        if (pitchCursor > segmentEnd) {
            pitchCursor = pitchCursor - segmentLen;
        }
        if (pitchCursor < segmentStart) {
            pitchCursor = pitchCursor + segmentLen;
        }
        terminationCursor = terminationCursor + 1;
    }
    // see if we went out of range, and if so, transpose in direction of proper
    // range (allowing for slight overshoot)
    rangeTestEnd = gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].cursor;
    minPitch = gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[rangeTestStart].pitch;
    maxPitch = minPitch;
    for (rangeTestCursor = rangeTestStart; rangeTestCursor < rangeTestEnd; rangeTestCursor = rangeTestCursor + 1) {
        rangePitch = gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[rangeTestCursor].pitch;
        if (rangePitch < minPitch) {
            minPitch = rangePitch;
        }
        if (rangePitch > maxPitch) {
            maxPitch = rangePitch;
        }
    }
    rangeMin = gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].rangeBottom;
    rangeMax = gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].rangeTop;
    if ((minPitch < rangeMin) && (maxPitch > rangeMax)) {
        jsabort('exceeded both min and max for range');
        // exceeded both min and max -- need to scale and transpose
        rangeScalar = (rangeMax - rangeMin) / (maxPitch - minPitch);
        rangeTranspose = ((maxPitch - minPitch) / 2) - ((rangeMax - rangeMin) / 2);
        for (ix = rangeTestStart; ix < rangeTestEnd; ix = ix + 1) {
            gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[ix].pitch = (gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[ix].pitch * rangeScalar) - rangeTranspose;
        }
    } else {
        if (minPitch < rangeMin) {
            rangeTranspose = (rangeMin - minPitch) / 2;
            for (ix = rangeTestStart; ix < rangeTestEnd; ix = ix + 1) {
                gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[ix].pitch = gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[ix].pitch + rangeTranspose;
            }
        } else {
            if (maxPitch > rangeMax) {
                rangeTranspose = (maxPitch - rangeMax) / 2;
                for (ix = rangeTestStart; ix < rangeTestEnd; ix = ix + 1) {
                    gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[ix].pitch = gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[ix].pitch - rangeTranspose;
                }
            }
        }
    }
}

function umtUiIsPartInUI(ptnum) {
    "use strict";
    if (document.getElementById("voice_row_" + ptnum)) {
        return true;
    }
    return false;
}

function umtComposeTab(lpnum, tab) {
    "use strict";
    var compositionVoice, UIVoice, songNumber, octave, rangeMin, rangeMax, instrumentName, percussion, exemptFromHarmonization, noRepeatNotes, relativeNotes, minSize, maxSize, chunkSize, totalLength, centerOctave, restyness, ampVariation, volume, skew, noteDistance, startTime, endTime, rhythmDirection, pitchesDirection, timeRhythmScroll, timePitchScroll, actualPitchScroll, pitchScale, pitchBevel, pitchTranspose, symmetryPattern, csrPos, csrStart, duration, tdvidx, currentTimeDiv, keepGoing, loopDuration, symSwitchPoint, sspidx, symPoint, vcnum, instParamNames, instSpecificParams, instFixedParams;
    if (gUmt.noReenterAddingVoice) {
        jsabort("umtComposeTab called during add voice");
    }
    cx("Composing tab " + ctstr(tab) + " to loop " + ctstr(lpnum));
    umtClearLoop(lpnum, tab, gUmt.UIParams.songTab[tab].tempo, gUmt.UIParams.songTab[tab].centernote, gUmt.UIParams.songTab[tab].scale, gUmt.UIParams.songTab[tab].chord);
    compositionVoice = 0;
    // calculate the loop duration
    loopDuration = 1;
    symSwitchPoint = []; // we keep track of the symmetry pattern switching
    // points to use below, since they now vary depending on time divisions
    sspidx = 0;
    keepGoing = true;
    for (tdvidx in gUmt.UIParams.timeDivisions) {
        if (gUmt.UIParams.timeDivisions.hasOwnProperty(tdvidx)) {
            currentTimeDiv = gUmt.UIParams.timeDivisions[tdvidx];
            if (currentTimeDiv === 1) {
                keepGoing = false;
            }
            if (keepGoing) {
                symSwitchPoint[sspidx] = loopDuration;
                loopDuration = loopDuration * currentTimeDiv;
                sspidx = sspidx + 1;
            }
        }
    }
    symSwitchPoint[sspidx] = loopDuration; // we tack on the last time, even
    // though it's not a symmetry pattern switch point, just so the code below
    // that checks for the end doesn't have to treat the end as a special case
    for (UIVoice = 0; UIVoice < gUmt.UIParams.parts; UIVoice = UIVoice + 1) {
        if (gUmt.UIParams.songTab[tab].voice[UIVoice].playing) {
            songNumber = gUmt.UIParams.songTab[tab].voice[UIVoice].songNumber;
            octave = gUmt.UIParams.songTab[tab].voice[UIVoice].octave;
            rangeMin = octave - 1.25;
            rangeMax = octave + 1.25;
            instrumentName = gUmt.UIParams.songTab[tab].voice[UIVoice].instrument;
            percussion = gUmt.instrumentParams[instrumentName].percussion;
            exemptFromHarmonization = gUmt.UIParams.songTab[tab].voice[UIVoice].exemptFromHarmonization;
            noRepeatNotes = gUmt.UIParams.songTab[tab].voice[UIVoice].noRepeatNotes;
            relativeNotes = gUmt.UIParams.songTab[tab].voice[UIVoice].relativeNotes;
            umtAddVoice(lpnum, tab, compositionVoice, songNumber, rangeMin, rangeMax, instrumentName, percussion, exemptFromHarmonization, noRepeatNotes, relativeNotes);
            minSize = gUmt.UIParams.songTab[tab].voice[UIVoice].minNote;
            maxSize = gUmt.UIParams.songTab[tab].voice[UIVoice].maxNote;
            chunkSize = gUmt.UIParams.songTab[tab].voice[UIVoice].chunkSize;
            totalLength = gUmt.UIParams.songTab[tab].voice[UIVoice].frameSize;
            centerOctave = octave;
            restyness = gUmt.UIParams.songTab[tab].voice[UIVoice].restyness;
            ampVariation = gUmt.UIParams.songTab[tab].voice[UIVoice].ampVariation;
            noteDistance = gUmt.UIParams.songTab[tab].voice[UIVoice].noteDistance;
            volume = gUmt.UIParams.songTab[tab].voice[UIVoice].volume;
            skew = gUmt.UIParams.songTab[tab].voice[UIVoice].skew;
            instParamNames = gUmt.instrumentParams[instrumentName].parameters;
            instSpecificParams = gUmt.UIParams.songTab[tab].voice[UIVoice].instSpecificParams;
            instFixedParams = gUmt.UIParams.songTab[tab].voice[UIVoice].instFixedParams;
            umtAddSymmetryBaseToVoice(lpnum, tab, compositionVoice, minSize, maxSize, chunkSize, totalLength, centerOctave, restyness, noteDistance, relativeNotes, ampVariation, volume, skew, instrumentName, instParamNames, instSpecificParams, instFixedParams);
            // we have voices and base rhythms; now need symmetry
            startTime = 0;
            csrPos = gUmt.loop[lpnum].score.songTab[tab].voice[compositionVoice].cursor;
            csrStart = gUmt.loop[lpnum].score.songTab[tab].voice[compositionVoice].notes[csrPos - 1].starttime; // can eliminate intermediaries
            duration = gUmt.loop[lpnum].score.songTab[tab].voice[compositionVoice].notes[csrPos - 1].duration;
            endTime = csrStart + duration;
            sspidx = 0;
            while (symSwitchPoint[sspidx] !== endTime) {
                sspidx = sspidx + 1;
            }
            symPoint = endTime;
            while (endTime < loopDuration) {
                // set default parameters, then modify defaults according to UI
                rhythmDirection = 1;
                pitchesDirection = 1;
                timeRhythmScroll = 0;
                timePitchScroll = 0;
                actualPitchScroll = 0;
                pitchScale = 1.0;
                pitchBevel = 0;
                pitchTranspose = 0;
                symmetryPattern = gUmt.UIParams.songTab[tab].voice[UIVoice].symmetry[symPoint].pattern;
                switch (symmetryPattern) {
                case 'trans':
                    pitchTranspose = (2 * gUmt.UIParams.songTab[tab].voice[UIVoice].symmetry[symPoint].transParam) * (gUmt.loop[lpnum].score.songTab[tab].voice[compositionVoice].randTrans.genrandReal1() - 0.5);
                    break;
                case 'invert':
                    pitchScale = -1.0;
                    break;
                case 'scale':
                    pitchScale = gUmt.UIParams.songTab[tab].voice[UIVoice].symmetry[symPoint].scaleParam * gUmt.loop[lpnum].score.songTab[tab].voice[compositionVoice].randScale.genrandReal1();
                    break;
                case 'tilt':
                    pitchBevel = gUmt.UIParams.songTab[tab].voice[UIVoice].symmetry[symPoint].tiltParam * gUmt.loop[lpnum].score.songTab[tab].voice[compositionVoice].randTilt.genrandReal1();
                    pitchTranspose = -(pitchBevel / 2);
                    break;
                case 'reverserhythm':
                    rhythmDirection = -1;
                    break;
                case 'reversepitch':
                    pitchesDirection = -1;
                    break;
                case 'reverseboth':
                    rhythmDirection = -1;
                    pitchesDirection = -1;
                    break;
                case 'scrollrhythm':
                    timeRhythmScroll = (gUmt.UIParams.songTab[tab].voice[UIVoice].symmetry[symPoint].scrollbothParam * gUmt.loop[lpnum].score.songTab[tab].voice[compositionVoice].randTilt.genrandReal1()) * (symPoint - startTime);
                    break;
                case 'scrollpitch':
                    timePitchScroll = (gUmt.UIParams.songTab[tab].voice[UIVoice].symmetry[symPoint].scrollbothParam * gUmt.loop[lpnum].score.songTab[tab].voice[compositionVoice].randTilt.genrandReal1()) * (symPoint - startTime);
                    break;
                case 'scrollboth':
                    timeRhythmScroll = (gUmt.UIParams.songTab[tab].voice[UIVoice].symmetry[symPoint].scrollbothParam * gUmt.loop[lpnum].score.songTab[tab].voice[compositionVoice].randTilt.genrandReal1()) * (symPoint - startTime);
                    timePitchScroll = timeRhythmScroll;
                    break;
                default:
                    jsabort("unrecognized symmetry pattern in composition: " + symmetryPattern);
                    break;
                }
                vcnum = compositionVoice;
                umtCopyAndTransform(lpnum, tab, vcnum, startTime, symPoint, rhythmDirection, pitchesDirection, timeRhythmScroll, timePitchScroll, actualPitchScroll, pitchScale, pitchBevel, pitchTranspose);
                startTime = 0;
                csrPos = gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].cursor;
                csrStart = gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[csrPos - 1].starttime;
                duration = gUmt.loop[lpnum].score.songTab[tab].voice[vcnum].notes[csrPos - 1].duration;
                endTime = csrStart + duration;
                if (endTime === symSwitchPoint[sspidx + 1]) {
                    sspidx = sspidx + 1;
                    symPoint = endTime;
                }
            }
            compositionVoice = compositionVoice + 1;
        }
    }
    umtHarmonizeScore(lpnum, tab);
}

function umtMakeSureTabRecomposeFlagsExist(lpnum, numTabs) {
    "use strict";
    var tab;
    if (!gUmt.tabRecomposeFlags.hasOwnProperty(lpnum)) {
        gUmt.tabRecomposeFlags[lpnum] = [];
    }
    for (tab = 0; tab < numTabs; tab = tab + 1) {
        if (!gUmt.tabRecomposeFlags[lpnum].hasOwnProperty(tab)) {
            gUmt.tabRecomposeFlags[lpnum][tab] = true; // default to true; tab is dirty, needs to be recomposed
        }
    }
}

function umtSetAllTabsAllLoopsToRecompose() {
    "use strict";
    var lpnum, tab, numTabs;
    numTabs = gUmt.UIParams.numTabs;
    for (lpnum = 0; lpnum < 2; lpnum = lpnum + 1) {
        if (!gUmt.tabRecomposeFlags.hasOwnProperty(lpnum)) {
            gUmt.tabRecomposeFlags[lpnum] = [];
        }
        for (tab = 0; tab < numTabs; tab = tab + 1) {
            gUmt.tabRecomposeFlags[lpnum][tab] = true; // default to true; tab is dirty, needs to be recomposed
        }
    }
}

function umtComposeAllTabs(lpnum) {
    "use strict";
    var tab, numTabs;
    numTabs = gUmt.UIParams.numTabs;
    umtMakeSureTabRecomposeFlagsExist(lpnum, numTabs);
    for (tab = 0; tab < numTabs; tab = tab + 1) {
        if (gUmt.tabRecomposeFlags[lpnum][tab]) {
            umtComposeTab(lpnum, tab);
        }
        gUmt.tabRecomposeFlags[lpnum][tab] = false;
    }
    // these must be done afterward because "score" object might not exist until after call to umtComposeTab()
    gUmt.loop[lpnum].score.numTabs = numTabs;
    gUmt.loop[lpnum].score.playAllTabs = true;
}

// composition algorithm demarcation

function umtExecTimerPulse() {
    "use strict";
    umtDoTimerPulse();
}

function umtExecAutomaticStart() {
    "use strict";
    var tab;
    if (gUmt.noReenterCompose === false) {
        gUmt.noReenterCompose = true;
        if (gUmt.UIParams.playAllTabs) {
            gUmt.tabRecomposeFlags = [];
            umtComposeAllTabs(1);
            gUmt.loop[1].score.playAllTabs = true;
        } else {
            tab = gUmt.UIParams.currentTab;
            umtComposeTab(1, tab);
            gUmt.loop[1].score.playAllTabs = false;
        }
    }
    umtAutomaticStart(1);
    gUmt.noReenterCompose = false;
}

function umtExecAutomaticStop() {
    "use strict";
    console.log(gUmt);
    umtAutomaticStop();
    umtPlaybackHighlightOneTab(-1);
}

function umtDisableUI() {
    "use strict";
    document.getElementById("start_automatic").disabled = true;
    document.getElementById("stop_automatic").disabled = true;
}

function umtEnableUI() {
    "use strict";
    document.getElementById("start_automatic").disabled = false;
    document.getElementById("stop_automatic").disabled = false;
}

function umtAutomaticallyRecomposeLoopFromUIParams() {
    "use strict";
    var oldLoop, newLoop, startTime, stopTime, diffTime, tab;
    cx("Automatically Recompose Loop From UI Params");
    if (gUmt.timerIsOn) {
        oldLoop = gUmt.currentlyPlayingLoop;
        newLoop = 1 - oldLoop;
    } else {
        return;
    }
    startTime = gUmt.globalCtx.currentTime;
    umtDisableUI();
    tab = gUmt.UIParams.currentTab;
    if (gUmt.UIParams.playAllTabs) {
        umtMakeSureTabRecomposeFlagsExist(0, gUmt.UIParams.numTabs);
        umtMakeSureTabRecomposeFlagsExist(1, gUmt.UIParams.numTabs);
        gUmt.tabRecomposeFlags[0][tab] = true;
        gUmt.tabRecomposeFlags[1][tab] = true;
        umtComposeAllTabs(newLoop);
        gUmt.loop[newLoop].score.playAllTabs = true;
        gUmt.loop[newLoop].score.numTabs = gUmt.UIParams.numTabs;
    } else {
        umtComposeTab(newLoop, tab);
        gUmt.loop[newLoop].score.playAllTabs = false;
    }
    umtAutomaticStart(newLoop);
    umtEnableUI();
    stopTime = gUmt.globalCtx.currentTime;
    diffTime = stopTime - startTime;
    if (diffTime > gUmt.compositionLag) {
        gUmt.compositionLag = diffTime;
    } else {
        gUmt.compositionLag = (0.9 * gUmt.compositionLag) + (0.1 * diffTime);
    }
    document.getElementById("test_composition_time").value = diffTime;
}

function umtExecScaleChange() {
    "use strict";
    var listbox, scaleparam, tab;
    listbox = document.getElementById("lb_scale");
    scaleparam = listbox.options[listbox.selectedIndex].value;
    tab = gUmt.UIParams.currentTab;
    gUmt.UIParams.songTab[tab].scale = scaleparam;
    if (gUmt.noReenterAddingVoice === false) {
        umtAutomaticallyRecomposeLoopFromUIParams();
    }
}

function umtExecHarmonizationChange() {
    "use strict";
    var listbox, chordparam, tab;
    listbox = document.getElementById("lb_harmonization");
    chordparam = listbox.options[listbox.selectedIndex].value;
    tab = gUmt.UIParams.currentTab;
    gUmt.UIParams.songTab[tab].chord = chordparam;
    if (gUmt.noReenterAddingVoice === false) {
        umtAutomaticallyRecomposeLoopFromUIParams();
    }
}

function umtExecPlayVoiceChange(event) {
    "use strict";
    var elemid, UIVoiceNum, checkd, tab;
    elemid = event.target.id;
    UIVoiceNum = Number(elemid.substring(9));
    checkd = document.getElementById(elemid).checked;
    tab = gUmt.UIParams.currentTab;
    gUmt.UIParams.songTab[tab].voice[UIVoiceNum].playing = checkd;
    if (gUmt.noReenterAddingVoice === false) {
        umtAutomaticallyRecomposeLoopFromUIParams();
    }
}

function umtExecSongNumberChange(event) {
    "use strict";
    var elemid, UIVoiceNum, listbox, tab;
    elemid = event.target.id;
    UIVoiceNum = Number(elemid.substring(10));
    listbox = document.getElementById("songnumber" + UIVoiceNum);
    tab = gUmt.UIParams.currentTab;
    gUmt.UIParams.songTab[tab].voice[UIVoiceNum].songNumber = Number(listbox.options[listbox.selectedIndex].value);
    if (gUmt.noReenterAddingVoice === false) {
        if (gUmt.UIParams.songTab[tab].voice[UIVoiceNum].playing) {
            umtAutomaticallyRecomposeLoopFromUIParams();
        }
    }
}

function umtUiCreateClosureFunctionsForSymmetrySlidersOutsideALoop(voiceNumber, patternSize) {
    "use strict";
    jQuery("#slider_symmetry_transParam" + voiceNumber + "_" + patternSize).slider({min: 0, max: 100, step: 5, value: 0, orientation: "horizontal", change: function (event, ui) {
        var tab, newvalue, patternDuration;
        tab = gUmt.UIParams.currentTab;
        // patternDuration and patternSize are two names for the same thing
        // but we have to use two different names so they will refer to
        // different variables and the closure will work
        patternDuration = patternSize;
        if (gUmt.noReenterAddingVoice === false) {
            if (event.type === "slidechange") {
                newvalue = ui.value / 100.0;
                document.getElementById("display_symmetry_transParam" + voiceNumber + "_" + patternDuration).innerHTML = Math.round(newvalue * 100) + "%";
                gUmt.UIParams.songTab[tab].voice[voiceNumber].symmetry[patternDuration].transParam = newvalue;
                if (gUmt.UIParams.songTab[tab].voice[voiceNumber].playing) {
                    umtAutomaticallyRecomposeLoopFromUIParams();
                }
            }
        }
    }
         });
    jQuery("#slider_symmetry_scaleParam" + voiceNumber + "_" + patternSize).slider({min: 0, max: 240, step: 5, value: 120, orientation: "horizontal", change: function (event, ui) {
        var tab, newvalue, patternDuration;
        tab = gUmt.UIParams.currentTab;
        patternDuration = patternSize;
        if (gUmt.noReenterAddingVoice === false) {
            if (event.type === "slidechange") {
                newvalue = (ui.value - 120.0) / 100.0;
                document.getElementById("display_symmetry_scaleParam" + voiceNumber + "_" + patternDuration).innerHTML = Math.round(newvalue * 100) + "%";
                gUmt.UIParams.songTab[tab].voice[voiceNumber].symmetry[patternDuration].scaleParam = newvalue;
                if (gUmt.UIParams.songTab[tab].voice[voiceNumber].playing) {
                    umtAutomaticallyRecomposeLoopFromUIParams();
                }
            }
        }
    }
         });
    jQuery("#slider_symmetry_tiltParam" + voiceNumber + "_" + patternSize).slider({min: 0, max: 240, step: 5, value: 120, orientation: "horizontal", change: function (event, ui) {
        var tab, newvalue, patternDuration;
        tab = gUmt.UIParams.currentTab;
        patternDuration = patternSize;
        if (gUmt.noReenterAddingVoice === false) {
            if (event.type === "slidechange") {
                newvalue = (ui.value - 120.0) / 100.0;
                document.getElementById("display_symmetry_tiltParam" + voiceNumber + "_" + patternDuration).innerHTML = Math.round(newvalue * 100) + "%";
                gUmt.UIParams.songTab[tab].voice[voiceNumber].symmetry[patternDuration].tiltParam = newvalue;
                if (gUmt.UIParams.songTab[tab].voice[voiceNumber].playing) {
                    umtAutomaticallyRecomposeLoopFromUIParams();
                }
            }
        }
    }
         });
    jQuery("#slider_symmetry_scrollrhythmParam" + voiceNumber + "_" + patternSize).slider({min: 0, max: 100, step: 5, value: 0, orientation: "horizontal", change: function (event, ui) {
        var tab, newvalue, patternDuration;
        tab = gUmt.UIParams.currentTab;
        patternDuration = patternSize;
        if (gUmt.noReenterAddingVoice === false) {
            if (event.type === "slidechange") {
                newvalue = ui.value / 100.0;
                document.getElementById("display_symmetry_scrollrhythmParam" + voiceNumber + "_" + patternDuration).innerHTML = Math.round(newvalue * 100) + "%";
                gUmt.UIParams.songTab[tab].voice[voiceNumber].symmetry[patternDuration].scrollrhythmParam = newvalue;
                if (gUmt.UIParams.songTab[tab].voice[voiceNumber].playing) {
                    umtAutomaticallyRecomposeLoopFromUIParams();
                }
            }
        }
    }
         });
    jQuery("#slider_symmetry_scrollpitchParam" + voiceNumber + "_" + patternSize).slider({min: 0, max: 100, step: 5, value: 0, orientation: "horizontal", change: function (event, ui) {
        var tab, newvalue, patternDuration;
        tab = gUmt.UIParams.currentTab;
        patternDuration = patternSize;
        if (gUmt.noReenterAddingVoice === false) {
            if (event.type === "slidechange") {
                newvalue = ui.value / 100.0;
                document.getElementById("display_symmetry_scrollpitchParam" + voiceNumber + "_" + patternDuration).innerHTML = Math.round(newvalue * 100) + "%";
                gUmt.UIParams.songTab[tab].voice[voiceNumber].symmetry[patternDuration].scrollpitchParam = newvalue;
                if (gUmt.UIParams.songTab[tab].voice[voiceNumber].playing) {
                    umtAutomaticallyRecomposeLoopFromUIParams();
                }
            }
        }
    }
         });
    jQuery("#slider_symmetry_scrollbothParam" + voiceNumber + "_" + patternSize).slider({min: 0, max: 100, step: 5, value: 0, orientation: "horizontal", change: function (event, ui) {
        var tab, newvalue, patternDuration;
        tab = gUmt.UIParams.currentTab;
        patternDuration = patternSize;
        if (gUmt.noReenterAddingVoice === false) {
            if (event.type === "slidechange") {
                newvalue = ui.value / 100.0;
                document.getElementById("display_symmetry_scrollbothParam" + voiceNumber + "_" + patternDuration).innerHTML = Math.round(newvalue * 100) + "%";
                gUmt.UIParams.songTab[tab].voice[voiceNumber].symmetry[patternDuration].scrollbothParam = newvalue;
                if (gUmt.UIParams.songTab[tab].voice[voiceNumber].playing) {
                    umtAutomaticallyRecomposeLoopFromUIParams();
                }
            }
        }
    }
         });
}

function umtUiShowOrHideAppropriateSymmetrySliders(symmetrySelected, suffix) {
    "use strict";
    var displayTrans, displayScale, displayTilt, displayScrollrhythm, displayScrollpitch, displayScrollboth;
    displayTrans = "none";
    displayScale = "none";
    displayTilt = "none";
    displayScrollrhythm = "none";
    displayScrollpitch = "none";
    displayScrollboth = "none";
    switch (symmetrySelected) {
    case 'trans':
        displayTrans = "block";
        break;
    case 'invert':
        break;
    case 'scale':
        displayScale = "block";
        break;
    case 'tilt':
        displayTilt = "block";
        break;
    case 'reverserhythm':
        break;
    case 'reversepitch':
        break;
    case 'reverseboth':
        break;
    case 'scrollrhythm':
        displayScrollrhythm = "block";
        break;
    case 'scrollpitch':
        displayScrollpitch = "block";
        break;
    case 'scrollboth':
        displayScrollboth = "block";
        break;
    default:
        jsabort("unrecognized symmetry selected: " + symmetrySelected);
        break;
    }
    document.getElementById("slider_symmetry_transParam" + suffix).style.display = displayTrans;
    document.getElementById("display_symmetry_transParam" + suffix).style.display = displayTrans;
    document.getElementById("slider_symmetry_scaleParam" + suffix).style.display = displayScale;
    document.getElementById("display_symmetry_scaleParam" + suffix).style.display = displayScale;
    document.getElementById("slider_symmetry_tiltParam" + suffix).style.display = displayTilt;
    document.getElementById("display_symmetry_tiltParam" + suffix).style.display = displayTilt;
    document.getElementById("slider_symmetry_scrollrhythmParam" + suffix).style.display = displayScrollrhythm;
    document.getElementById("display_symmetry_scrollrhythmParam" + suffix).style.display = displayScrollrhythm;
    document.getElementById("slider_symmetry_scrollpitchParam" + suffix).style.display = displayScrollpitch;
    document.getElementById("display_symmetry_scrollpitchParam" + suffix).style.display = displayScrollpitch;
    document.getElementById("slider_symmetry_scrollbothParam" + suffix).style.display = displayScrollboth;
    document.getElementById("display_symmetry_scrollbothParam" + suffix).style.display = displayScrollboth;
}

function umtExecPervoiceSymmetryChange(event) {
    "use strict";
    var elemid, listbox, selectedvalue, i, voiceNumber, symSize, symmetryPattern, tab;
    elemid = event.target.id;
    listbox = event.target;
    selectedvalue = listbox.options[listbox.selectedIndex].value;
    elemid = elemid.substring(12);
    umtUiShowOrHideAppropriateSymmetrySliders(selectedvalue, elemid);
    i = elemid.indexOf("_");
    voiceNumber = elemid.substring(0, i);
    symSize = elemid.substring(i + 1);
    symmetryPattern = selectedvalue;
    tab = gUmt.UIParams.currentTab;
    gUmt.UIParams.songTab[tab].voice[voiceNumber].symmetry[symSize].pattern = symmetryPattern;
    if (gUmt.UIParams.songTab[tab].voice[voiceNumber].playing) {
        umtAutomaticallyRecomposeLoopFromUIParams();
    }
}

function umtUiTurnOnSymmetryPatterns(row, patternLabel, isOn) {
    "use strict";
    var suffix, block, inline;
    suffix = row + "_" + patternLabel;
    if (isOn) {
        block = "block";
        inline = "inline";
    } else {
        block = "none";
        inline = "none";
    }
    document.getElementById("display_symmetrylevel" + suffix).style.display = inline;
    document.getElementById("display_symmetrylevel" + suffix).innerHTML = patternLabel;
    document.getElementById("sel_symmetry" + suffix).style.display = inline;
    document.getElementById("slider_symmetry_transParam" + suffix).style.display = block;
    document.getElementById("display_symmetry_transParam" + suffix).style.display = block;
    document.getElementById("slider_symmetry_scaleParam" + suffix).style.display = block;
    document.getElementById("display_symmetry_scaleParam" + suffix).style.display = block;
    document.getElementById("slider_symmetry_tiltParam" + suffix).style.display = block;
    document.getElementById("display_symmetry_tiltParam" + suffix).style.display = block;
    document.getElementById("slider_symmetry_scrollrhythmParam" + suffix).style.display = block;
    document.getElementById("display_symmetry_scrollrhythmParam" + suffix).style.display = block;
    document.getElementById("slider_symmetry_scrollpitchParam" + suffix).style.display = block;
    document.getElementById("display_symmetry_scrollpitchParam" + suffix).style.display = block;
    document.getElementById("slider_symmetry_scrollbothParam" + suffix).style.display = block;
    document.getElementById("display_symmetry_scrollbothParam" + suffix).style.display = block;
}

function umtUiSetSymmetryListboxAndSlider(row, patternDuration) {
    "use strict";
    var suffix, absvalue, slidervalue, tab;
    suffix = row + "_" + patternDuration;
    tab = gUmt.UIParams.currentTab;
    document.getElementById("sel_symmetry" + suffix).value = gUmt.UIParams.songTab[tab].voice[row].symmetry[patternDuration].pattern;
    // trans param
    absvalue = gUmt.UIParams.songTab[tab].voice[row].symmetry[patternDuration].transParam;
    slidervalue = Math.floor(absvalue * 100);
    document.getElementById("display_symmetry_transParam" + suffix).innerHTML = Math.round(absvalue * 100) + "%";
    jQuery("#slider_symmetry_transParam" + suffix).slider({ value: slidervalue });
    // scale param
    absvalue = gUmt.UIParams.songTab[tab].voice[row].symmetry[patternDuration].scaleParam;
    slidervalue = Math.floor((absvalue + 1.0) * 100);
    document.getElementById("display_symmetry_scaleParam" + suffix).innerHTML = Math.round(absvalue * 100) + "%";
    jQuery("#slider_symmetry_scaleParam" + suffix).slider({ value: slidervalue });
    // tilt param
    absvalue = gUmt.UIParams.songTab[tab].voice[row].symmetry[patternDuration].tiltParam;
    slidervalue = Math.floor((absvalue + 1.0) * 100);
    document.getElementById("display_symmetry_tiltParam" + suffix).innerHTML = Math.round(absvalue * 100) + "%";
    jQuery("#slider_symmetry_tiltParam" + suffix).slider({ value: slidervalue });
    // scrollrhythm param
    absvalue = gUmt.UIParams.songTab[tab].voice[row].symmetry[patternDuration].scrollrhythmParam;
    slidervalue = absvalue * 100;
    document.getElementById("display_symmetry_scrollrhythmParam" + suffix).innerHTML = Math.round(absvalue * 100) + "%";
    jQuery("#slider_symmetry_scrollrhythmParam" + suffix).slider({ value: slidervalue });
    // scrollpitch param
    absvalue = gUmt.UIParams.songTab[tab].voice[row].symmetry[patternDuration].scrollpitchParam;
    slidervalue = absvalue * 100;
    document.getElementById("display_symmetry_scrollpitchParam" + suffix).innerHTML = Math.round(absvalue * 100) + "%";
    jQuery("#slider_symmetry_scrollpitchParam" + suffix).slider({ value: slidervalue });
    // scrollboth param
    absvalue = gUmt.UIParams.songTab[tab].voice[row].symmetry[patternDuration].scrollbothParam;
    slidervalue = absvalue * 100;
    document.getElementById("display_symmetry_scrollbothParam" + suffix).innerHTML = Math.round(absvalue * 100) + "%";
    jQuery("#slider_symmetry_scrollbothParam" + suffix).slider({ value: slidervalue });
}

function umtUiTimeDivMultToIdx(value) {
    "use strict";
    var tdvidx, count, thisTimeDiv, mult;
    count = 0;
    mult = 1;
    if (mult === value) {
        return count;
    }
    for (tdvidx in gUmt.UIParams.timeDivisions) {
        if (gUmt.UIParams.timeDivisions.hasOwnProperty(tdvidx)) {
            thisTimeDiv = gUmt.UIParams.timeDivisions[tdvidx];
            if (thisTimeDiv === 1) {
                return count;
            }
            mult = mult * thisTimeDiv;
            count = count + 1;
            if (mult === value) {
                return count;
            }
        }
    }
    return count;
}

function umtUiGetSymmetryAlreadyThere(voiceNumber) {
    "use strict";
    var alreadySym; // name has to be fixed because this is actually a return
    // value; we're using a closure to get a return value out of a callback
    // function in order to avoid using a global variable.
    alreadySym = [];
    jQuery('#voicesymmetriescell' + voiceNumber).children().each(function () {
        var elemid, symSize, i;
        elemid = this.id;
        if (elemid.substring(0, 12) === "tbl_symmetry") {
            elemid = elemid.substring(12);
            i = elemid.indexOf("_");
            if (i > 0) {
                symSize = Number(elemid.substring(i + 1));
                alreadySym.push(symSize);
            }
        }
    });
    return alreadySym;
}

function umtUiSetListboxByValue(listbox, value) {
    "use strict";
    var options, i, valstr;
    options = listbox.options;
    valstr = value.toString();
    for (i = 0; i < options.length; i = i + 1) {
        if (options[i].value === valstr) {
            options[i].selected = true;
        }
    }
}

function umtUiSetSymmetrySliders(tab, voiceNumber, frameSize) {
    "use strict";
    var keepGoing, mult, tdvidx, currentTimeDiv, alreadySym, alreadyIdx, alreadyOffEnd, match, symmetryPattern;
    keepGoing = true;
    mult = 1;
    alreadySym = umtUiGetSymmetryAlreadyThere(voiceNumber);
    alreadyIdx = 0;
    if (alreadySym.length === 0) {
        alreadyOffEnd = true;
    } else {
        alreadyOffEnd = false;
    }
    for (tdvidx in gUmt.UIParams.timeDivisions) {
        if (gUmt.UIParams.timeDivisions.hasOwnProperty(tdvidx)) {
            currentTimeDiv = gUmt.UIParams.timeDivisions[tdvidx];
            if (currentTimeDiv === 1) {
                keepGoing = false;
            }
            if (keepGoing) {
                match = false;
                while ((!alreadyOffEnd) && (mult > alreadySym[alreadyIdx])) {
                    umtUiTurnOnSymmetryPatterns(voiceNumber, alreadySym[alreadyIdx], false);
                    alreadyIdx = alreadyIdx + 1;
                    if (alreadyIdx === alreadySym.length) {
                        alreadyOffEnd = true;
                    }
                }
                if (!alreadyOffEnd) {
                    if (mult === alreadySym[alreadyIdx]) {
                        match = true;
                        alreadyIdx = alreadyIdx + 1; // we put this here so that
                        // the +1 8 lines above will catch entries that need to
                        // be shut off
                        if (alreadyIdx === alreadySym.length) {
                            alreadyOffEnd = true;
                        }
                    }
                }
                if (match) {
                    if (mult < frameSize) {
                        umtUiTurnOnSymmetryPatterns(voiceNumber, mult, false);
                    } else {
                        umtUiTurnOnSymmetryPatterns(voiceNumber, mult, true);
                        umtUiSetSymmetryListboxAndSlider(voiceNumber, mult);
                        symmetryPattern = gUmt.UIParams.songTab[tab].voice[voiceNumber].symmetry[mult].pattern;
                        // symmetry listbox
                        umtUiSetListboxByValue(document.getElementById("sel_symmetry" + voiceNumber + "_" + mult), symmetryPattern);
                        umtUiShowOrHideAppropriateSymmetrySliders(symmetryPattern, voiceNumber + "_" + mult);
                    }
                } else {
                    umtUiTurnOnSymmetryPatterns(voiceNumber, mult, false);
                }
                mult = mult * currentTimeDiv;
            }
        }
    }
    while (!alreadyOffEnd) {
        umtUiTurnOnSymmetryPatterns(voiceNumber, alreadySym[alreadyIdx], false);
        alreadyIdx = alreadyIdx + 1;
        if (alreadyIdx === alreadySym.length) {
            alreadyOffEnd = true;
        }
    }
}

var umtUiCascadeFrameSizeChanges, umtUiCascadeChunkSizeChanges, umtUiCascadeMaxNoteChanges, umtUiCascadeMinNoteChanges;

umtUiCascadeFrameSizeChanges = function (voiceNumber, newvalue) {
    "use strict";
    var tab, chunkSize, slidervalue;
    tab = gUmt.UIParams.currentTab;
    chunkSize = gUmt.UIParams.songTab[tab].voice[voiceNumber].chunkSize;
    if (newvalue < chunkSize) {
        chunkSize = newvalue;
        // slidervalue = Math.floor((Math.log(chunkSize) / Math.log(2)) + 0.01);
        slidervalue = umtUiTimeDivMultToIdx(chunkSize);
        jQuery("#slider_chunksize" + voiceNumber).slider({ value: slidervalue });
        document.getElementById("display_chunksize" + voiceNumber).innerHTML = newvalue;
        gUmt.UIParams.songTab[tab].voice[voiceNumber].chunkSize = chunkSize;
        // cx("voiceNumber" + ctstr(voiceNumber) + " chunkSize" + ctstr(chunkSize));
        umtUiCascadeChunkSizeChanges(voiceNumber, chunkSize);
    }
    umtUiSetSymmetrySliders(tab, voiceNumber, newvalue);
};

umtUiCascadeChunkSizeChanges = function (voiceNumber, newvalue) {
    "use strict";
    var tab, maxNote, frameSize, slidervalue;
    tab = gUmt.UIParams.currentTab;
    maxNote = gUmt.UIParams.songTab[tab].voice[voiceNumber].maxNote;
    if (newvalue < maxNote) {
        maxNote = newvalue;
        // slidervalue = Math.floor((Math.log(maxNote) / Math.log(2)) + 0.01);
        slidervalue = umtUiTimeDivMultToIdx(maxNote);
        jQuery("#slider_maxnote" + voiceNumber).slider({ value: slidervalue });
        document.getElementById("display_maxnote" + voiceNumber).innerHTML = newvalue;
        gUmt.UIParams.songTab[tab].voice[voiceNumber].maxNote = maxNote;
        umtUiCascadeMaxNoteChanges(voiceNumber, maxNote);
    }
    frameSize = gUmt.UIParams.songTab[tab].voice[voiceNumber].frameSize;
    if (newvalue > frameSize) {
        frameSize = newvalue;
        // slidervalue = Math.floor((Math.log(frameSize) / Math.log(2)) + 0.01);
        slidervalue = umtUiTimeDivMultToIdx(frameSize);
        jQuery("#slider_framesize" + voiceNumber).slider({ value: slidervalue });
        document.getElementById("display_framesize" + voiceNumber).innerHTML = newvalue;
        gUmt.UIParams.songTab[tab].voice[voiceNumber].frameSize = frameSize;
        umtUiCascadeFrameSizeChanges(voiceNumber, frameSize);
    }
};

umtUiCascadeMaxNoteChanges = function (voiceNumber, newvalue) {
    "use strict";
    var tab, minNote, chunkSize, slidervalue;
    tab = gUmt.UIParams.currentTab;
    minNote = gUmt.UIParams.songTab[tab].voice[voiceNumber].minNote;
    if (newvalue < minNote) {
        minNote = newvalue;
        // slidervalue = Math.floor((Math.log(minNote) / Math.log(2)) + 0.01);
        slidervalue = umtUiTimeDivMultToIdx(minNote);
        jQuery("#slider_minnote" + voiceNumber).slider({ value: slidervalue });
        document.getElementById("display_minnote" + voiceNumber).innerHTML = newvalue;
        gUmt.UIParams.songTab[tab].voice[voiceNumber].minNote = minNote;
        umtUiCascadeMinNoteChanges(voiceNumber, minNote);
    }
    chunkSize = gUmt.UIParams.songTab[tab].voice[voiceNumber].chunkSize;
    if (newvalue > chunkSize) {
        chunkSize = newvalue;
        // slidervalue = Math.floor((Math.log(chunkSize) / Math.log(2)) + 0.01);
        slidervalue = umtUiTimeDivMultToIdx(chunkSize);
        jQuery("#slider_chunksize" + voiceNumber).slider({ value: slidervalue });
        document.getElementById("display_chunksize" + voiceNumber).innerHTML = newvalue;
        gUmt.UIParams.songTab[tab].voice[voiceNumber].chunkSize = chunkSize;
        umtUiCascadeChunkSizeChanges(voiceNumber, chunkSize);
    }
};

umtUiCascadeMinNoteChanges = function (voiceNumber, newvalue) {
    "use strict";
    var tab, maxNote, slidervalue;
    tab = gUmt.UIParams.currentTab;
    maxNote = gUmt.UIParams.songTab[tab].voice[voiceNumber].maxNote;
    if (newvalue > maxNote) {
        maxNote = newvalue;
        // slidervalue = Math.floor((Math.log(maxNote) / Math.log(2)) + 0.01);
        slidervalue = umtUiTimeDivMultToIdx(maxNote);
        jQuery("#slider_maxnote" + voiceNumber).slider({ value: slidervalue });
        document.getElementById("display_maxnote" + voiceNumber).innerHTML = newvalue;
        gUmt.UIParams.songTab[tab].voice[voiceNumber].maxNote = maxNote;
        umtUiCascadeMaxNoteChanges(voiceNumber, maxNote);
    }
};

function umtUiSetDefaultInstSpecificParamsForNewInstrument(voiceNumber, instrumentName) {
    "use strict";
    var tab, instParameters, instParamIdx, instParamSet, paramName;
    tab = gUmt.UIParams.currentTab;
    instParameters = gUmt.instrumentParams[instrumentName].parameters;
    for (instParamIdx in instParameters) {
        if (instParameters.hasOwnProperty(instParamIdx)) {
            instParamSet = instParameters[instParamIdx];
            paramName = instParamSet.name;
            if (!gUmt.UIParams.songTab[tab].voice[voiceNumber].instSpecificParams.hasOwnProperty(instrumentName + "_" + paramName + "_patterns")) {
                gUmt.UIParams.songTab[tab].voice[voiceNumber].instSpecificParams[instrumentName + "_" + paramName + "_patterns"] = 0.0;
            }
        }
    }
}

function umtUiCreateClosureFunctionsForInstrumentSpecificParametersOutsideALoop(instrumentName, paramName, voiceNumber) {
    "use strict";
    jQuery("#slider_instr_" + instrumentName + "_" + paramName + "_" + voiceNumber + "_patterns").slider({min: 0, max: 100, step: 5, value: 0, orientation: "vertical", change: function (event, ui) {
        var tab, newvalue, absvalue, aspercentage;
        if (gUmt.noReenterAddingVoice === false) {
            if (event.type === "slidechange") {
                tab = gUmt.UIParams.currentTab;
                newvalue = ui.value;
                absvalue = newvalue / 100.0;
                aspercentage = newvalue;
                document.getElementById("display_instr_" + instrumentName + "_" + paramName + "_" + voiceNumber + "_patterns").innerHTML = Math.round(aspercentage) + "%";
                gUmt.UIParams.songTab[tab].voice[voiceNumber].instSpecificParams[instrumentName + "_" + paramName + "_patterns"] = absvalue;
                if (gUmt.UIParams.songTab[tab].voice[voiceNumber].playing) {
                    umtAutomaticallyRecomposeLoopFromUIParams();
                }
            }
        }
    }
        });
}

function umtCreateInstrumentSpecificParameterControls(voiceNumber, instrumentName) {
    "use strict";
    var cellidname, text, i, instrSpecificControlHTML, parameters, idx, instpar, template;
    cellidname = "instparameterscell";
    text = document.getElementById(cellidname + "xx").innerHTML;
    i = text.indexOf("xx");
    while (i >= 0) {
        text = text.substring(0, i) + voiceNumber + text.substring(i + 2);
        i = text.indexOf("xx");
    }
    instrSpecificControlHTML = '<table border="0"><tr>';
    parameters = gUmt.instrumentParams[instrumentName].parameters;
    for (idx in parameters) {
        if (parameters.hasOwnProperty(idx)) {
            instpar = parameters[idx];
            instrSpecificControlHTML = instrSpecificControlHTML + '<td>' + instpar.display + '<br />';
            template = document.getElementById("instrument_parameter_template").innerHTML;
            i = template.indexOf("ZZ");
            while (i >= 0) {
                template = template.substring(0, i) + instrumentName + template.substring(i + 2);
                i = template.indexOf("ZZ");
            }
            i = template.indexOf("YY");
            while (i >= 0) {
                template = template.substring(0, i) + instpar.name + template.substring(i + 2);
                i = template.indexOf("YY");
            }
            i = template.indexOf("XX");
            while (i >= 0) {
                template = template.substring(0, i) + voiceNumber + template.substring(i + 2);
                i = template.indexOf("XX");
            }
            instrSpecificControlHTML = instrSpecificControlHTML + template + '</td>';
        }
    }
    instrSpecificControlHTML = instrSpecificControlHTML + '</tr></table>';
    i = text.indexOf("<!-- inst_specific_param_ins_point -->");
    text = text.substring(0, i) + instrSpecificControlHTML + text.substring(i); // paste
    document.getElementById(cellidname + voiceNumber).innerHTML = text;
    for (idx in parameters) {
        if (parameters.hasOwnProperty(idx)) {
            instpar = parameters[idx];
            umtUiCreateClosureFunctionsForInstrumentSpecificParametersOutsideALoop(instrumentName, instpar.name, voiceNumber);
        }
    }
}

function umtUiSetInstrumentSpecificSliders(voiceNumber) {
    "use strict";
    var tab, instrumentName, instParameters, instParamIdx, instParamSet, paramName, absvalue, slidervalue;
    tab = gUmt.UIParams.currentTab;
    instrumentName = gUmt.UIParams.songTab[tab].voice[voiceNumber].instrument;
    instParameters = gUmt.instrumentParams[instrumentName].parameters;
    for (instParamIdx in instParameters) {
        if (instParameters.hasOwnProperty(instParamIdx)) {
            instParamSet = instParameters[instParamIdx];
            paramName = instParamSet.name;
            absvalue = gUmt.UIParams.songTab[tab].voice[voiceNumber].instSpecificParams[instrumentName + "_" + paramName + "_patterns"];
            slidervalue = absvalue * 100;
            document.getElementById("display_instr_" + instrumentName + "_" + paramName + "_" + voiceNumber + "_patterns").innerHTML = Math.round(absvalue * 100) + "%";
            jQuery("#slider_instr_" + instrumentName + "_" + paramName + "_" + voiceNumber + "_patterns").slider({ value: slidervalue });
        }
    }
}

function umtUiSetDefaultInstFixedParamsForNewInstrument(voiceNumber, instrumentName) {
    "use strict";
    var tab, instFixed, instFixedIdx, instFixedSet, fixedName;
    tab = gUmt.UIParams.currentTab;
    instFixed = gUmt.instrumentParams[instrumentName].fixed;
    for (instFixedIdx in instFixed) {
        if (instFixed.hasOwnProperty(instFixedIdx)) {
            instFixedSet = instFixed[instFixedIdx];
            fixedName = instFixedSet.name;
            if (!gUmt.UIParams.songTab[tab].voice[voiceNumber].instFixedParams.hasOwnProperty(instrumentName + "_" + fixedName)) {
                gUmt.UIParams.songTab[tab].voice[voiceNumber].instFixedParams[instrumentName + "_" + fixedName] = instFixedSet.default;
            }
        }
    }
}

function umtExecPervoiceFixedChange(event) {
    "use strict";
    var elemid, listbox, selectedvalue, i, voiceNum, tab, paramName, instrumentName;
    tab = gUmt.UIParams.currentTab;
    elemid = event.target.id;
    // what user selected
    listbox = event.target;
    selectedvalue = listbox.options[listbox.selectedIndex].value;
    // remove choose_fixed_
    elemid = elemid.substring(13);
    // instrument name
    i = elemid.indexOf("_");
    instrumentName = elemid.substring(0, i);
    elemid = elemid.substring(i + 1);
    // param name
    i = elemid.indexOf("_");
    paramName = elemid.substring(0, i);
    // voice number
    voiceNum = Number(elemid.substring(i + 1));
    // put it all together and set the parameter
    gUmt.UIParams.songTab[tab].voice[voiceNum].instFixedParams[instrumentName + "_" + paramName] = selectedvalue;
    if (gUmt.UIParams.songTab[tab].voice[voiceNum].playing) {
        umtAutomaticallyRecomposeLoopFromUIParams();
    }
}

function umtUiCreateClosureFunctionsForInstrumentFixedParametersOutsideALoop(instrumentName, paramName, voiceNumber) {
    "use strict";
    var listenElement;
    listenElement = document.getElementById("choose_fixed_" + instrumentName + "_" + paramName + "_" + voiceNumber);
    listenElement.addEventListener("change", umtExecPervoiceFixedChange, true);
}

function umtCreateInstrumentFixedParameterControls(voiceNumber, instrumentName) {
    "use strict";
    var cellidname, text, i, instrFixedControlHTML, instFixed, instFixedIdx, instFixedSet, template, fixedName, optionText, optionIdx, optionList, optionStr, optionCode, optionDisplay, osidx;
    cellidname = "fixedparams";
    text = document.getElementById(cellidname + "xx").innerHTML;
    i = text.indexOf("xx"); // this should actually be nothing, but included for consistency with instrument-specific params
    while (i >= 0) {
        text = text.substring(0, i) + voiceNumber + text.substring(i + 2);
        i = text.indexOf("xx");
    }
    instrFixedControlHTML = '';
    instFixed = gUmt.instrumentParams[instrumentName].fixed;
    for (instFixedIdx in instFixed) {
        if (instFixed.hasOwnProperty(instFixedIdx)) {
            instFixedSet = instFixed[instFixedIdx];
            fixedName = instFixedSet.name;
            instrFixedControlHTML = instrFixedControlHTML + instFixedSet.display + '<br />';
            template = document.getElementById("instrument_fixed_param_template").innerHTML;
            i = template.indexOf("ZZ");
            while (i >= 0) {
                template = template.substring(0, i) + instrumentName + template.substring(i + 2);
                i = template.indexOf("ZZ");
            }
            i = template.indexOf("YY");
            while (i >= 0) {
                template = template.substring(0, i) + fixedName + template.substring(i + 2);
                i = template.indexOf("YY");
            }
            i = template.indexOf("XX");
            while (i >= 0) {
                template = template.substring(0, i) + voiceNumber + template.substring(i + 2);
                i = template.indexOf("XX");
            }
            i = template.indexOf("inst_fixed_param_listbox_options_ins_point");
            if (i >= 0) {
                optionText = "";
                optionList = instFixedSet.values;
                for (optionIdx in optionList) {
                    if (optionList.hasOwnProperty(optionIdx)) {
                        optionStr = optionList[optionIdx];
                        osidx = optionStr.indexOf("=");
                        optionCode = optionStr.substring(0, osidx);
                        optionDisplay = optionStr.substring(osidx + 1);
                        optionText = optionText + '<option value="' + optionCode + '">' + optionDisplay + '</option>';
                    }
                }
                template = template.substring(0, i - 5) + optionText + template.substring(i + 42 + 4); // -5 and +4 to remove the HTML comment tags
            }
            instrFixedControlHTML = instrFixedControlHTML + template + '<br />';
        }
    }
    document.getElementById(cellidname + voiceNumber).innerHTML = instrFixedControlHTML;
    for (instFixedIdx in instFixed) {
        if (instFixed.hasOwnProperty(instFixedIdx)) {
            instFixedSet = instFixed[instFixedIdx];
            umtUiCreateClosureFunctionsForInstrumentFixedParametersOutsideALoop(instrumentName, instFixedSet.name, voiceNumber);
        }
    }
}

function umtUiSetInstrumentFixedParams(voiceNumber) {
    "use strict";
    var tab, instrumentName, instFixed, instFixedIdx, instFixedSet, paramName, paramValue;
    tab = gUmt.UIParams.currentTab;
    instrumentName = gUmt.UIParams.songTab[tab].voice[voiceNumber].instrument;
    instFixed = gUmt.instrumentParams[instrumentName].fixed;
    for (instFixedIdx in instFixed) {
        if (instFixed.hasOwnProperty(instFixedIdx)) {
            instFixedSet = instFixed[instFixedIdx];
            paramName = instFixedSet.name;
            paramValue = gUmt.UIParams.songTab[tab].voice[voiceNumber].instFixedParams[instrumentName + "_" + paramName];
            umtUiSetListboxByValue(document.getElementById("choose_fixed_" + instrumentName + "_" + paramName + "_" + voiceNumber), paramValue);
        }
    }
}

function umtExecInstrumentChange() {
    "use strict";
    var elemid, UIVoiceNum, listbox, newinstr, tab;
    elemid = event.target.id;
    UIVoiceNum = Number(elemid.substring(10));
    listbox = document.getElementById("instrument" + UIVoiceNum);
    newinstr = listbox.options[listbox.selectedIndex].value;
    tab = gUmt.UIParams.currentTab;
    gUmt.UIParams.songTab[tab].voice[UIVoiceNum].instrument = newinstr;
    umtUiSetDefaultInstFixedParamsForNewInstrument(UIVoiceNum, newinstr);
    umtUiSetDefaultInstSpecificParamsForNewInstrument(UIVoiceNum, newinstr);
    umtCreateInstrumentFixedParameterControls(UIVoiceNum, newinstr);
    umtCreateInstrumentSpecificParameterControls(UIVoiceNum, newinstr);
    umtUiSetInstrumentSpecificSliders(UIVoiceNum);
    if (gUmt.noReenterAddingVoice === false) {
        umtAutomaticallyRecomposeLoopFromUIParams();
    }
}

function umtUiCopySymmetry(fromTab, fromVoice, fromSymSize, toTab, toVoice, toSymSize) {
    "use strict";
    var pvsym, newPattern;
    pvsym = gUmt.UIParams.songTab[fromTab].voice[fromVoice].symmetry;
    newPattern = {
        pattern: pvsym[fromSymSize].pattern,
        transParam: pvsym[fromSymSize].transParam,
        scaleParam: pvsym[fromSymSize].scaleParam,
        tiltParam: pvsym[fromSymSize].tiltParam,
        scrollrhythmParam: pvsym[fromSymSize].scrollrhythmParam,
        scrollpitchParam: pvsym[fromSymSize].scrollpitchParam,
        scrollbothParam: pvsym[fromSymSize].scrollbothParam
    };
    gUmt.UIParams.songTab[toTab].voice[toVoice].symmetry[toSymSize] = newPattern;
}

// universal voice copy function -- used by both Add Voice and Add Tab
function umtUiCopyTabVoice(fromTab, fromVoice, toTab, toVoice) {
    "use strict";
    var prev, pvinstspec, currentinstspec, pvinstfixed, currentinstfixed, instParamName, mult, keepGoing, tdvidx, currentTimeDiv;
    prev = gUmt.UIParams.songTab[fromTab].voice[fromVoice];
    gUmt.UIParams.songTab[toTab].voice[toVoice] = {
        playing: prev.playing,
        exemptFromHarmonization: prev.exemptFromHarmonization,
        instrument: prev.instrument,
        volume: prev.volume,
        songNumber: prev.songNumber,
        octave: prev.octave,
        noteDistance: prev.noteDistance,
        noRepeatNotes: prev.noRepeatNotes,
        relativeNotes: prev.relativeNotes,
        restyness: prev.restyness,
        ampVariation: prev.ampVariation,
        skew: prev.skew,
        minNote: prev.minNote,
        maxNote: prev.maxNote,
        chunkSize: prev.chunkSize,
        frameSize: prev.frameSize,
        symmetry: {},
        instSpecificParams: {},
        instFixedParams: {},
        copyChangesToOtherTabs: false
    };
    mult = 1;
    keepGoing = true;
    for (tdvidx in gUmt.UIParams.timeDivisions) {
        if (gUmt.UIParams.timeDivisions.hasOwnProperty(tdvidx)) {
            currentTimeDiv = gUmt.UIParams.timeDivisions[tdvidx];
            if (currentTimeDiv === 1) {
                keepGoing = false;
            }
            if (keepGoing) {
                umtUiCopySymmetry(fromTab, fromVoice, mult, toTab, toVoice, mult);
                mult = mult * currentTimeDiv;
            }
        }
    }
    // instrument-specifc
    pvinstspec = gUmt.UIParams.songTab[fromTab].voice[fromVoice].instSpecificParams;
    currentinstspec = gUmt.UIParams.songTab[toTab].voice[toVoice].instSpecificParams;
    for (instParamName in pvinstspec) {
        if (pvinstspec.hasOwnProperty(instParamName)) {
            currentinstspec[instParamName] = pvinstspec[instParamName];
        }
    }
    // instrument-fixed
    pvinstfixed = gUmt.UIParams.songTab[fromTab].voice[fromVoice].instFixedParams;
    currentinstfixed = gUmt.UIParams.songTab[toTab].voice[toVoice].instFixedParams;
    for (instParamName in pvinstfixed) {
        if (pvinstfixed.hasOwnProperty(instParamName)) {
            currentinstfixed[instParamName] = pvinstfixed[instParamName];
        }
    }
}

function umtExecVoiceExemptFromHarmonizationChange(event) {
    "use strict";
    var elemid, UIVoiceNum, checkd, tab;
    elemid = event.target.id;
    UIVoiceNum = Number(elemid.substring(28));
    checkd = document.getElementById(elemid).checked;
    tab = gUmt.UIParams.currentTab;
    gUmt.UIParams.songTab[tab].voice[UIVoiceNum].exemptFromHarmonization = checkd;
    if (gUmt.noReenterAddingVoice === false) {
        umtAutomaticallyRecomposeLoopFromUIParams();
    }
}

function umtExecVoiceNoRepeatNotesChange(event) {
    "use strict";
    var elemid, UIVoiceNum, checkd, tab;
    elemid = event.target.id;
    UIVoiceNum = Number(elemid.substring(18));
    checkd = document.getElementById(elemid).checked;
    tab = gUmt.UIParams.currentTab;
    gUmt.UIParams.songTab[tab].voice[UIVoiceNum].noRepeatNotes = checkd;
    if (gUmt.noReenterAddingVoice === false) {
        umtAutomaticallyRecomposeLoopFromUIParams();
    }
}

function umtExecVoiceRelativeNotesChange(event) {
    "use strict";
    var elemid, UIVoiceNum, checkd, tab;
    elemid = event.target.id;
    UIVoiceNum = Number(elemid.substring(18));
    checkd = document.getElementById(elemid).checked;
    tab = gUmt.UIParams.currentTab;
    gUmt.UIParams.songTab[tab].voice[UIVoiceNum].relativeNotes = checkd;
    if (gUmt.noReenterAddingVoice === false) {
        umtAutomaticallyRecomposeLoopFromUIParams();
    }
}

var umtExecDeleteVoiceClick; // has to exist for next function; can't be rearranged because tried that and functions really are circular

function umtUiTimeDivIdxToMult(tdvidx) {
    "use strict";
    var mult, i, thisTimeDiv;
    mult = 1;
    for (i = 0; i < tdvidx; i = i + 1) {
        if (gUmt.UIParams.timeDivisions.hasOwnProperty(i)) {
            thisTimeDiv = gUmt.UIParams.timeDivisions[i];
            if (thisTimeDiv === 1) {
                return mult;
            }
            mult = mult * gUmt.UIParams.timeDivisions[i];
        }
    }
    return mult;
}

function umtUiInsertSymmetrySliders(voiceNumber) {
    "use strict";
    var keepGoing, mult, tdvidx, currentTimeDiv, alreadySym, alreadyIdx, alreadyOffEnd, needToInsert, text, i;
    keepGoing = true;
    mult = 1;
    alreadySym = umtUiGetSymmetryAlreadyThere(voiceNumber);
    alreadyIdx = 0;
    if (alreadySym.length === 0) {
        alreadyOffEnd = true;
    } else {
        alreadyOffEnd = false;
    }
    for (tdvidx in gUmt.UIParams.timeDivisions) {
        if (gUmt.UIParams.timeDivisions.hasOwnProperty(tdvidx)) {
            currentTimeDiv = gUmt.UIParams.timeDivisions[tdvidx];
            if (currentTimeDiv === 1) {
                keepGoing = false;
            }
            if (keepGoing) {
                needToInsert = true;
                while ((!alreadyOffEnd) && (mult > alreadySym[alreadyIdx])) {
                    alreadyIdx = alreadyIdx + 1;
                    if (alreadyIdx === alreadySym.length) {
                        alreadyOffEnd = true;
                    }
                }
                if (!alreadyOffEnd) {
                    if (mult === alreadySym[alreadyIdx]) {
                        needToInsert = false;
                    }
                }
                if (needToInsert) {
                    text = document.getElementById("symmetry_variations_template").innerHTML;
                    i = text.indexOf("xx");
                    while (i >= 0) {
                        text = text.substring(0, i) + voiceNumber + text.substring(i + 2);
                        i = text.indexOf("xx");
                    }
                    i = text.indexOf("yyy");
                    while (i >= 0) {
                        text = text.substring(0, i) + mult + text.substring(i + 3);
                        i = text.indexOf("yyy");
                    }
                    if (alreadyOffEnd) {
                        jQuery('#voicesymmetriescell' + voiceNumber).append(text);
                    } else {
                        jQuery('#tbl_symmetry' + voiceNumber + '_' + alreadySym[alreadyIdx]).before(text);
                    }
                }
                mult = mult * currentTimeDiv;
            }
        }
    }
}

function umtUiInsertAllSymmetrySliders() {
    "use strict";
    var parts, vcnum;
    parts = gUmt.UIParams.parts;
    for (vcnum = 0; vcnum < parts; vcnum = vcnum + 1) {
        umtUiInsertSymmetrySliders(vcnum);
    }
}

function umtUiAttachSymmetryEventListenerForVoice(voiceNumber) {
    "use strict";
    var mult, tdvidx, keepGoing, currentTimeDiv, listenElement;
    mult = 1;
    keepGoing = true;
    for (tdvidx in gUmt.UIParams.timeDivisions) {
        if (gUmt.UIParams.timeDivisions.hasOwnProperty(tdvidx)) {
            currentTimeDiv = gUmt.UIParams.timeDivisions[tdvidx];
            if (currentTimeDiv === 1) {
                keepGoing = false;
            }
            if (keepGoing) {
                listenElement = document.getElementById("sel_symmetry" + voiceNumber + "_" + mult);
                listenElement.addEventListener("change", umtExecPervoiceSymmetryChange, true);
                umtUiCreateClosureFunctionsForSymmetrySlidersOutsideALoop(voiceNumber, mult);
                mult = mult * currentTimeDiv;
            }
        }
    }
}

function umtUiAttachSymmetryEventListeners() {
    "use strict";
    var vcnum, parts;
    parts = gUmt.UIParams.parts;
    for (vcnum = 0; vcnum < parts; vcnum = vcnum + 1) {
        umtUiAttachSymmetryEventListenerForVoice(vcnum);
    }
}

function umtUiAddRow(voiceNumber) {
    "use strict";
    var text, i, displayNumber, columnidlist, idx, cellidname, listenElement;
    // First, add the HTML
    jQuery('#voices_table > tbody:last').append('<tr id="voice_row_' + voiceNumber + '"><td id="voicedescriptcell' + voiceNumber + '" valign="top"> Voice Description Cell ' + voiceNumber + '</td><td id="voiceplaycell' + voiceNumber + '" valign="top" align="center"></td><td id="voiceinstrumentcell' + voiceNumber + '" valign="top" align="center"><td id="voicevolumecell' + voiceNumber + '" valign="top" align="center"></td><td id="voiceexemptharmonizationcell' + voiceNumber + '" valign="top" align="center"></td><td id="voicesongnumcell' + voiceNumber + '" valign="top" align="center"></td><td id="voiceoctavecell' + voiceNumber + '" valign="top" align="center"></td><td id="voicenotedistcell' + voiceNumber + '" valign="top" align="center"></td><td id="voicenorepeatcell' + voiceNumber + '" valign="top" align="center"></td><td id="voicerestynesscell' + voiceNumber + '" valign="top" align="center"></td><td id="voiceampvariationcell' + voiceNumber + '" valign="top" align="center"></td><td id="voiceskewcell' + voiceNumber + '" valign="top" align="center"></td><td id="voiceminnotecell' + voiceNumber + '" valign="top" align="center"><td id="voicemaxnotecell' + voiceNumber + '" valign="top" align="center"></td><td id="voicechunksizecell' + voiceNumber + '" valign="top" align="center"></td><td id="voiceframesizecell' + voiceNumber + '" valign="top" align="center"></td><td id="voicesymmetriescell' + voiceNumber + '" valign="top" align="center"></td><td id="instparameterscell' + voiceNumber + '" valign="top" align="center"></td><td id="voicedeletecell' + voiceNumber + '" valign="top" align="center"></td></tr>');
    // Then, paste in standard templates for each of the cells
    text = document.getElementById("voicedescriptcellxx").innerHTML;
    displayNumber = voiceNumber + 1;
    i = text.indexOf("XX");
    text = text.substring(0, i) + displayNumber + text.substring(i + 2);
    document.getElementById("voicedescriptcell" + voiceNumber).innerHTML = text;
    columnidlist = ["voiceplaycell", "voiceinstrumentcell", "voicevolumecell", "voiceexemptharmonizationcell", "voicesongnumcell", "voiceoctavecell", "voicenotedistcell", "voicenorepeatcell", "voicerestynesscell", "voiceampvariationcell", "voiceskewcell", "voiceminnotecell", "voicemaxnotecell", "voicechunksizecell", "voiceframesizecell", "voicesymmetriescell", "instparameterscell", "voicedeletecell"];
    for (idx in columnidlist) {
        if (columnidlist.hasOwnProperty(idx)) {
            cellidname = columnidlist[idx];
            text = document.getElementById(cellidname + "xx").innerHTML;
            i = text.indexOf("xx");
            while (i >= 0) {
                text = text.substring(0, i) + voiceNumber + text.substring(i + 2);
                i = text.indexOf("xx");
            }
            document.getElementById(cellidname + voiceNumber).innerHTML = text;
        }
    }
    umtUiInsertSymmetrySliders(voiceNumber);
    // Add event listeners
    // Play
    listenElement = document.getElementById("playvoice" + voiceNumber);
    listenElement.addEventListener("change", umtExecPlayVoiceChange, true);
    // Instrument
    listenElement = document.getElementById("instrument" + voiceNumber);
    listenElement.addEventListener("change", umtExecInstrumentChange, true);
    // Exempt from harmonization
    listenElement = document.getElementById("voiceexemptfromharmonization" + voiceNumber);
    listenElement.addEventListener("change", umtExecVoiceExemptFromHarmonizationChange, true);
    // Song Number
    listenElement = document.getElementById("songnumber" + voiceNumber);
    listenElement.addEventListener("change", umtExecSongNumberChange, true);
    // No Repeat & Relative Note Distance
    listenElement = document.getElementById("voicenorepeatnotes" + voiceNumber);
    listenElement.addEventListener("change", umtExecVoiceNoRepeatNotesChange, true);
    listenElement = document.getElementById("voicerelativenotes" + voiceNumber);
    listenElement.addEventListener("change", umtExecVoiceRelativeNotesChange, true);
    // Volume
    jQuery("#slider_volume" + voiceNumber).slider({min: 0, max: 100, step: 5, value: 0, orientation: "vertical", change: function (event, ui) {
        var tab, newvalue, absvalue, aspercentage;
        tab = gUmt.UIParams.currentTab;
        if (gUmt.noReenterAddingVoice === false) {
            if (event.type === "slidechange") {
                newvalue = ui.value;
                absvalue = newvalue / 100.0;
                aspercentage = newvalue;
                document.getElementById("display_volume" + voiceNumber).innerHTML = Math.round(aspercentage) + "%";
                gUmt.UIParams.songTab[tab].voice[voiceNumber].volume = absvalue;
                if (gUmt.UIParams.songTab[tab].voice[voiceNumber].playing) {
                    umtAutomaticallyRecomposeLoopFromUIParams();
                }
            }
        }
    }
         });
    // Octave
    jQuery("#slider_octave" + voiceNumber).slider({min: 0, max: 9, step: 1, value: 5, orientation: "vertical", change: function (event, ui) {
        var tab, newvalue;
        tab = gUmt.UIParams.currentTab;
        if (gUmt.noReenterAddingVoice === false) {
            if (event.type === "slidechange") {
                newvalue = ui.value;
                document.getElementById("display_octave" + voiceNumber).innerHTML = newvalue;
                gUmt.UIParams.songTab[tab].voice[voiceNumber].octave = newvalue;
                if (gUmt.UIParams.songTab[tab].voice[voiceNumber].playing) {
                    umtAutomaticallyRecomposeLoopFromUIParams();
                }
            }
        }
    }
         });
    // Note distance
    jQuery("#slider_notedistance" + voiceNumber).slider({min: 0, max: 100, step: 5, value: 0, orientation: "vertical", change: function (event, ui) {
        var tab, newvalue, absvalue, aspercentage;
        tab = gUmt.UIParams.currentTab;
        if (gUmt.noReenterAddingVoice === false) {
            if (event.type === "slidechange") {
                newvalue = ui.value;
                absvalue = newvalue / 100.0;
                aspercentage = newvalue;
                document.getElementById("display_notedistance" + voiceNumber).innerHTML = Math.round(aspercentage) + "%";
                gUmt.UIParams.songTab[tab].voice[voiceNumber].noteDistance = absvalue;
                if (gUmt.UIParams.songTab[tab].voice[voiceNumber].playing) {
                    umtAutomaticallyRecomposeLoopFromUIParams();
                }
            }
        }
    }
         });
    // Restyness
    jQuery("#slider_restyness" + voiceNumber).slider({min: 0, max: 100, step: 5, value: 0, orientation: "vertical", change: function (event, ui) {
        var tab, newvalue, absvalue, aspercentage;
        tab = gUmt.UIParams.currentTab;
        if (gUmt.noReenterAddingVoice === false) {
            if (event.type === "slidechange") {
                newvalue = ui.value;
                absvalue = newvalue / 100.0;
                aspercentage = newvalue;
                document.getElementById("display_restyness" + voiceNumber).innerHTML = Math.round(aspercentage) + "%";
                gUmt.UIParams.songTab[tab].voice[voiceNumber].restyness = absvalue;
                if (gUmt.UIParams.songTab[tab].voice[voiceNumber].playing) {
                    umtAutomaticallyRecomposeLoopFromUIParams();
                }
            }
        }
    }
         });
    // Amplitude variation
    jQuery("#slider_ampvariation" + voiceNumber).slider({min: 0, max: 100, step: 5, value: 0, orientation: "vertical", change: function (event, ui) {
        var tab, newvalue, absvalue, aspercentage;
        tab = gUmt.UIParams.currentTab;
        if (gUmt.noReenterAddingVoice === false) {
            if (event.type === "slidechange") {
                newvalue = ui.value;
                absvalue = newvalue / 100.0;
                aspercentage = newvalue;
                document.getElementById("display_ampvariation" + voiceNumber).innerHTML = Math.round(aspercentage) + "%";
                gUmt.UIParams.songTab[tab].voice[voiceNumber].ampVariation = absvalue;
                if (gUmt.UIParams.songTab[tab].voice[voiceNumber].playing) {
                    umtAutomaticallyRecomposeLoopFromUIParams();
                }
            }
        }
    }
         });
    // Skew
    jQuery("#slider_skew" + voiceNumber).slider({min: 0, max: 100, step: 5, value: 0, orientation: "vertical", change: function (event, ui) {
        var tab, newvalue, absvalue, aspercentage;
        tab = gUmt.UIParams.currentTab;
        if (gUmt.noReenterAddingVoice === false) {
            if (event.type === "slidechange") {
                newvalue = ui.value;
                absvalue = newvalue / 100.0;
                aspercentage = newvalue;
                document.getElementById("display_skew" + voiceNumber).innerHTML = Math.round(aspercentage) + "%";
                gUmt.UIParams.songTab[tab].voice[voiceNumber].skew = absvalue;
                if (gUmt.UIParams.songTab[tab].voice[voiceNumber].playing) {
                    umtAutomaticallyRecomposeLoopFromUIParams();
                }
            }
        }
    }
         });
    // Min note
    jQuery("#slider_minnote" + voiceNumber).slider({min: 0, max: 7, step: 1, value: 3, orientation: "vertical", change: function (event, ui) {
        var tab, newvalue;
        tab = gUmt.UIParams.currentTab;
        if (gUmt.noReenterAddingVoice === false) {
            if (event.type === "slidechange") {
                newvalue = ui.value;
                // newvalue = Math.pow(2, newvalue);
                newvalue = umtUiTimeDivIdxToMult(newvalue);
                document.getElementById("display_minnote" + voiceNumber).innerHTML = newvalue;
                gUmt.UIParams.songTab[tab].voice[voiceNumber].minNote = newvalue;
                gUmt.noReenterAddingVoice = true;
                umtUiCascadeMinNoteChanges(voiceNumber, newvalue);
                gUmt.noReenterAddingVoice = false;
                if (gUmt.UIParams.songTab[tab].voice[voiceNumber].playing) {
                    umtAutomaticallyRecomposeLoopFromUIParams();
                }
            }
        }
    }
         });
    // Max note
    jQuery("#slider_maxnote" + voiceNumber).slider({min: 0, max: 7, step: 1, value: 3, orientation: "vertical", change: function (event, ui) {
        var tab, newvalue;
        tab = gUmt.UIParams.currentTab;
        if (gUmt.noReenterAddingVoice === false) {
            if (event.type === "slidechange") {
                newvalue = ui.value;
                // newvalue = Math.pow(2, newvalue);
                newvalue = umtUiTimeDivIdxToMult(newvalue);
                document.getElementById("display_maxnote" + voiceNumber).innerHTML = newvalue;
                gUmt.UIParams.songTab[tab].voice[voiceNumber].maxNote = newvalue;
                gUmt.noReenterAddingVoice = true;
                umtUiCascadeMaxNoteChanges(voiceNumber, newvalue);
                gUmt.noReenterAddingVoice = false;
                if (gUmt.UIParams.songTab[tab].voice[voiceNumber].playing) {
                    umtAutomaticallyRecomposeLoopFromUIParams();
                }
            }
        }
    }
         });
    // Chunk size
    jQuery("#slider_chunksize" + voiceNumber).slider({min: 0, max: 7, step: 1, value: 3, orientation: "vertical", change: function (event, ui) {
        var tab, newvalue;
        tab = gUmt.UIParams.currentTab;
        if (gUmt.noReenterAddingVoice === false) {
            if (event.type === "slidechange") {
                newvalue = ui.value;
                // newvalue = Math.pow(2, newvalue);
                newvalue = umtUiTimeDivIdxToMult(newvalue);
                document.getElementById("display_chunksize" + voiceNumber).innerHTML = newvalue;
                gUmt.UIParams.songTab[tab].voice[voiceNumber].chunkSize = newvalue;
                gUmt.noReenterAddingVoice = true;
                umtUiCascadeChunkSizeChanges(voiceNumber, newvalue);
                gUmt.noReenterAddingVoice = false;
                if (gUmt.UIParams.songTab[tab].voice[voiceNumber].playing) {
                    umtAutomaticallyRecomposeLoopFromUIParams();
                }
            }
        }
    }
         });
    // Frame size
    jQuery("#slider_framesize" + voiceNumber).slider({min: 0, max: 7, step: 1, value: 3, orientation: "vertical", change: function (event, ui) {
        // BUGBUG: Important -- the max value needs to me MANUALLY sync'd with the loop size -- this needs to be fixed to be computed!
        var tab, newvalue;
        tab = gUmt.UIParams.currentTab;
        if (gUmt.noReenterAddingVoice === false) {
            if (event.type === "slidechange") {
                newvalue = ui.value;
                // newvalue = Math.pow(2, newvalue);
                newvalue = umtUiTimeDivIdxToMult(newvalue);
                document.getElementById("display_framesize" + voiceNumber).innerHTML = newvalue;
                gUmt.UIParams.songTab[tab].voice[voiceNumber].frameSize = newvalue;
                gUmt.noReenterAddingVoice = true;
                umtUiCascadeFrameSizeChanges(voiceNumber, newvalue);
                gUmt.noReenterAddingVoice = false;
                if (gUmt.UIParams.songTab[tab].voice[voiceNumber].playing) {
                    umtAutomaticallyRecomposeLoopFromUIParams();
                }
            }
        }
    }
         });
    // Symmetries
    umtUiAttachSymmetryEventListenerForVoice(voiceNumber);
    // for (y = 0; y < 10; y = y + 1) {
    //     yyy = Math.pow(2, y);
    //     listenElement = document.getElementById("sel_symmetry" + voiceNumber + "_" + yyy);
    //     listenElement.addEventListener("change", umtExecPervoiceSymmetryChange, true);
    //     umtUiCreateClosureFunctionsForSymmetrySlidersOutsideALoop(voiceNumber, yyy);
    // }
    // Close box
    listenElement = document.getElementById("deletevoice" + voiceNumber);
    listenElement.addEventListener("click", umtExecDeleteVoiceClick, true);
}

function umtUiTimeDivisionListboxExists(tdvidx) {
    "use strict";
    if (document.getElementById("lb_time_division_" + tdvidx)) {
        return true;
    }
    return false;
}

var umtExecTimeDivisionChange;

function umtUiCreateTimeDivisionListbox(tdvidx) {
    "use strict";
    var text, i, listenElement;
    // grab template, stick in our time division index number (tdvidx),
    // addpend to the div that has the list of time division listboxes
    text = document.getElementById("time_division_template").innerHTML;
    i = text.indexOf("xx");
    text = text.substring(0, i) + tdvidx + text.substring(i + 2);
    jQuery('#time_division_list').append(text);
    // Add event listeners
    listenElement = document.getElementById("lb_time_division_" + tdvidx);
    listenElement.addEventListener("change", umtExecTimeDivisionChange, true);
}

function umtUiDestroyTimeDivisionListbox(tdvidx) {
    "use strict";
    jQuery("#lb_time_division_" + tdvidx).remove();
}

function umtUiSetTimeDivisionListbox(tdvidx, choice) {
    "use strict";
    umtUiSetListboxByValue(document.getElementById("lb_time_division_" + tdvidx), choice);
}

function umtUiSetTimeDivisionListboxes() {
    "use strict";
    // the idea here is to use *either* a "1" or going off the end of the list as
    // the end of the list, and to put a listbox with "1" (which will display the text "stop")
    // at the end in either case. If a "1" occurs in the list, we set the "pastend" variable
    var count, tdvidx, choice, pastend;
    count = 0;
    pastend = false;
    for (tdvidx in gUmt.UIParams.timeDivisions) {
        if (gUmt.UIParams.timeDivisions.hasOwnProperty(tdvidx)) {
            count = count + 1;
            choice = gUmt.UIParams.timeDivisions[tdvidx];
            if (pastend) {
                if (umtUiTimeDivisionListboxExists(tdvidx)) {
                    umtUiDestroyTimeDivisionListbox(tdvidx);
                }
            } else {
                if (!umtUiTimeDivisionListboxExists(tdvidx)) {
                    umtUiCreateTimeDivisionListbox(tdvidx);
                    umtUiSetTimeDivisionListbox(tdvidx, choice);
                }
                if (choice === 1) {
                    pastend = true;
                }
            }
        }
    }
    if (!pastend) {
        if (!umtUiTimeDivisionListboxExists(count)) {
            umtUiCreateTimeDivisionListbox(count);
            umtUiSetTimeDivisionListbox(count, 1);
        }
        count = count + 1;
    }
    while (umtUiTimeDivisionListboxExists(count)) {
        umtUiDestroyTimeDivisionListbox(count);
        count = count + 1;
    }
}

function umtUiGetTimeDivisionCount() {
    "use strict";
    var tdvidx, count;
    count = 0;
    for (tdvidx in gUmt.UIParams.timeDivisions) {
        if (gUmt.UIParams.timeDivisions.hasOwnProperty(tdvidx)) {
            if (gUmt.UIParams.timeDivisions[tdvidx] === 1) {
                return count;
            }
            count = count + 1;
        }
    }
    return count;
}

function umtTestUiGetTimeDivisionCount() {
    "use strict";
    var result;
    cx("checking umtTestUiGetTimeDivisionCount");
    gUmt.UIParams.timeDivisions = [ 2, 5, 3, 7, 13, 11 ];
    result = umtUiGetTimeDivisionCount();
    if (result !== 6) {
        jsabort("umtUiGetTimeDivisionCount test failed, should have been 6.");
    }
    gUmt.UIParams.timeDivisions = [ 2, 5, 3, 1, 13, 11 ];
    result = umtUiGetTimeDivisionCount();
    if (result !== 3) {
        jsabort("umtUiGetTimeDivisionCount test failed, should have been 3.");
    }
}

function umtUiGetTimeDivisionCount() {
    "use strict";
    var tdvidx, count;
    count = 0;
    for (tdvidx in gUmt.UIParams.timeDivisions) {
        if (gUmt.UIParams.timeDivisions.hasOwnProperty(tdvidx)) {
            if (gUmt.UIParams.timeDivisions[tdvidx] === 1) {
                return count;
            }
            count = count + 1;
        }
    }
    return count;
}

function umtUiCopyUiParamsToActualUi() {
    "use strict";
    var parts, tab, vcnum, absvalue, slidervalue, timeDivCount, frameSize;
    // Important note: This function does not copy the tabs! It copies all the values outside the tabs, and all the values inside the tab *for the current tab* but does not change or configure the tabs themselves!!
    gUmt.noReenterAddingVoice = true;
    // master volume
    absvalue = gUmt.UIParams.mastervol;
    document.getElementById("display_mastervol_value").innerHTML = (absvalue * 100) + "%";
    slidervalue = absvalue * 100;
    jQuery("#slider_mastervol").slider({ value: slidervalue });
    // play selected tab or all tabs (loop tabs "one" or "all")
    if (gUmt.UIParams.playAllTabs) {
        document.getElementById("loop_tabs_all").checked = true;
    } else {
        document.getElementById("loop_tabs_one").checked = true;
    }
    // Time divisions
    umtUiSetTimeDivisionListboxes();
    // current tab
    tab = gUmt.UIParams.currentTab;
    jQuery("#tabs").tabs("option", "active", tab);
    // scale
    umtUiSetListboxByValue(document.getElementById("lb_scale"), gUmt.UIParams.songTab[tab].scale);
    // chords
    umtUiSetListboxByValue(document.getElementById("lb_harmonization"), gUmt.UIParams.songTab[tab].chord);
    // center note
    absvalue = gUmt.UIParams.songTab[tab].centernote;
    document.getElementById("display_centernote_value").innerHTML = absvalue;
    slidervalue = (Math.log(absvalue / gUmt.calibration) / Math.log(5 / 4)) + 5.0;
    slidervalue = Math.floor(slidervalue + 0.1); // in case of rounding error on the logarithm
    jQuery("#slider_centernote").slider({ value: slidervalue });
    // tempo
    absvalue = gUmt.UIParams.songTab[tab].tempo;
    document.getElementById("display_tempo_value").innerHTML = absvalue;
    slidervalue = absvalue; // no translation for tempo!
    jQuery("#slider_tempo").slider({ value: slidervalue });
    // now we loop through each part
    parts = gUmt.UIParams.parts;
    for (vcnum = 0; vcnum < parts; vcnum = vcnum + 1) {
        if (!umtUiIsPartInUI(vcnum)) {
            umtUiAddRow(vcnum);
        }
    }
    for (vcnum = 0; vcnum < parts; vcnum = vcnum + 1) {
        // play voice checkbox
        if (gUmt.UIParams.songTab[tab].voice[vcnum].playing) {
            document.getElementById("playvoice" + vcnum).checked = true;
        } else {
            document.getElementById("playvoice" + vcnum).checked = false;
        }
        // exempt from harmonization
        if (gUmt.UIParams.songTab[tab].voice[vcnum].exemptFromHarmonization) {
            document.getElementById("voiceexemptfromharmonization" + vcnum).checked = true;
        } else {
            document.getElementById("voiceexemptfromharmonization" + vcnum).checked = false;
        }
        // instrument
        umtUiSetListboxByValue(document.getElementById("instrument" + vcnum), gUmt.UIParams.songTab[tab].voice[vcnum].instrument);
        // song number
        umtUiSetListboxByValue(document.getElementById("songnumber" + vcnum), gUmt.UIParams.songTab[tab].voice[vcnum].songNumber);
        // volume
        absvalue = gUmt.UIParams.songTab[tab].voice[vcnum].volume;
        document.getElementById("display_volume" + vcnum).innerHTML = Math.round(absvalue * 100) + "%";
        slidervalue = Math.floor(absvalue * 100);
        jQuery("#slider_volume" + vcnum).slider({ value: slidervalue });
        // octave
        absvalue = gUmt.UIParams.songTab[tab].voice[vcnum].octave;
        document.getElementById("display_octave" + vcnum).innerHTML = absvalue;
        slidervalue = absvalue;
        jQuery("#slider_octave" + vcnum).slider({ value: slidervalue });
        // note distance
        absvalue = gUmt.UIParams.songTab[tab].voice[vcnum].noteDistance;
        document.getElementById("display_notedistance" + vcnum).innerHTML = Math.round(absvalue * 100) + "%";
        slidervalue = Math.floor(absvalue * 100);
        jQuery("#slider_notedistance" + vcnum).slider({ value: slidervalue });
        // No repeat and relative note distance
        if (gUmt.UIParams.songTab[tab].voice[vcnum].noRepeatNotes) {
            document.getElementById("voicenorepeatnotes" + vcnum).checked = true;
        } else {
            document.getElementById("voicenorepeatnotes" + vcnum).checked = false;
        }
        if (gUmt.UIParams.songTab[tab].voice[vcnum].relativeNotes) {
            document.getElementById("voicerelativenotes" + vcnum).checked = true;
        } else {
            document.getElementById("voicerelativenotes" + vcnum).checked = false;
        }
        // restyness
        absvalue = gUmt.UIParams.songTab[tab].voice[vcnum].restyness;
        document.getElementById("display_restyness" + vcnum).innerHTML = Math.round(absvalue * 100) + "%";
        slidervalue = Math.floor(absvalue * 100);
        jQuery("#slider_restyness" + vcnum).slider({ value: slidervalue });
        // amplitude variation
        absvalue = gUmt.UIParams.songTab[tab].voice[vcnum].ampVariation;
        document.getElementById("display_ampvariation" + vcnum).innerHTML = Math.round(absvalue * 100) + "%";
        slidervalue = Math.floor(absvalue * 100);
        jQuery("#slider_ampvariation" + vcnum).slider({ value: slidervalue });
        // skew
        absvalue = gUmt.UIParams.songTab[tab].voice[vcnum].skew;
        document.getElementById("display_skew" + vcnum).innerHTML = Math.round(absvalue * 100) + "%";
        slidervalue = Math.floor(absvalue * 100);
        jQuery("#slider_skew" + vcnum).slider({ value: slidervalue });
        // min note
        timeDivCount = umtUiGetTimeDivisionCount();
        jQuery("#slider_minnote" + vcnum).slider({ max: timeDivCount });
        absvalue = gUmt.UIParams.songTab[tab].voice[vcnum].minNote;
        document.getElementById("display_minnote" + vcnum).innerHTML = absvalue;
        // slidervalue = Math.floor((Math.log(absvalue) / Math.log(2)) + 0.01);
        slidervalue = umtUiTimeDivMultToIdx(absvalue);
        jQuery("#slider_minnote" + vcnum).slider({ value: slidervalue });
        // max note
        jQuery("#slider_maxnote" + vcnum).slider({ max: timeDivCount });
        absvalue = gUmt.UIParams.songTab[tab].voice[vcnum].maxNote;
        document.getElementById("display_maxnote" + vcnum).innerHTML = absvalue;
        // slidervalue = Math.floor((Math.log(absvalue) / Math.log(2)) + 0.01);
        slidervalue = umtUiTimeDivMultToIdx(absvalue);
        jQuery("#slider_maxnote" + vcnum).slider({ value: slidervalue });
        // chunk size
        jQuery("#slider_chunksize" + vcnum).slider({ max: timeDivCount });
        absvalue = gUmt.UIParams.songTab[tab].voice[vcnum].chunkSize;
        document.getElementById("display_chunksize" + vcnum).innerHTML = absvalue;
        // slidervalue = Math.floor((Math.log(absvalue) / Math.log(2)) + 0.01);
        slidervalue = umtUiTimeDivMultToIdx(absvalue);
        jQuery("#slider_chunksize" + vcnum).slider({ value: slidervalue });
        // frame size
        jQuery("#slider_framesize" + vcnum).slider({ max: timeDivCount });
        absvalue = gUmt.UIParams.songTab[tab].voice[vcnum].frameSize;
        document.getElementById("display_framesize" + vcnum).innerHTML = absvalue;
        // slidervalue = Math.floor((Math.log(absvalue) / Math.log(2)) + 0.01);
        slidervalue = umtUiTimeDivMultToIdx(absvalue);
        jQuery("#slider_framesize" + vcnum).slider({ value: slidervalue });
        // fixed but instrument-specific parameters
        umtCreateInstrumentFixedParameterControls(vcnum, gUmt.UIParams.songTab[tab].voice[vcnum].instrument);
        umtUiSetInstrumentFixedParams(vcnum);
        // symmetry patterns and parameters
        umtCreateInstrumentSpecificParameterControls(vcnum, gUmt.UIParams.songTab[tab].voice[vcnum].instrument);
        umtUiSetInstrumentSpecificSliders(vcnum);
        frameSize = gUmt.UIParams.songTab[tab].voice[vcnum].frameSize;
        umtUiSetSymmetrySliders(tab, vcnum, frameSize);
    }
    vcnum = parts;
    while (umtUiIsPartInUI(vcnum)) {
        jQuery("#voice_row_" + vcnum).remove();
        vcnum = vcnum + 1;
    }
    gUmt.noReenterAddingVoice = false;
}

function umtMakeTimeDivMultipliedList() {
    "use strict";
    var tdvidx, choice, pastend, count, mult, result;
    count = 0;
    mult = 1;
    result = [ 1 ];
    pastend = false;
    for (tdvidx in gUmt.UIParams.timeDivisions) {
        if (gUmt.UIParams.timeDivisions.hasOwnProperty(tdvidx)) {
            choice = gUmt.UIParams.timeDivisions[tdvidx];
            if (choice === 1) {
                pastend = true;
            }
            if (!pastend) {
                count = count + 1;
                mult = mult * choice;
                result[count] = mult;
            }
        }
    }
    return result;
}

function umtUiFindInTimeDivMultList(multList, value) {
    "use strict";
    var idx;
    for (idx in multList) {
        if (multList.hasOwnProperty(idx)) {
            if (multList[idx] === value) {
                return idx;
            }
        }
    }
    // We should never reach this point. If we do, blow up with an error.
    jsabort("We tried to find a value in a time division multiples list but the value was not found");
}

function umtUiFindNewTimeDivisionValue(before, after, value) {
    "use strict";
    var idx;
    idx = umtUiFindInTimeDivMultList(before, value);
    if (after.hasOwnProperty(idx)) {
        return after[idx];
    }
    idx = idx - 1; // there's an invisible typecast from string to int on this statement
    while (!after.hasOwnProperty(idx)) {
        idx = idx - 1;
    }
    return after[idx];
}

function umtUiRecalculateAllTimeDivisions(tdvidx, choice) {
    "use strict";
    // the idea here is to make one list of all the multiplications of the time
    // divisions from *before* the choice is changed, then change it, then make
    // a 2nd list, then use the two lists to "fix" all the rhythm (min note, max
    // note, chunk size, frame size) values, then the entire UI is redrawn,
    // fixing both those and the symmetry pattern section.
    var before, after, tab, vcnum, bidx, mapFrom, mapTo;
    before = umtMakeTimeDivMultipliedList();
    gUmt.UIParams.timeDivisions[tdvidx] = choice;
    after = umtMakeTimeDivMultipliedList();
    for (tab in gUmt.UIParams.songTab) {
        if (gUmt.UIParams.songTab.hasOwnProperty(tab)) {
            for (vcnum in gUmt.UIParams.songTab[tab].voice) {
                if (gUmt.UIParams.songTab[tab].voice.hasOwnProperty(vcnum)) {
                    gUmt.UIParams.songTab[tab].voice[vcnum].minNote = umtUiFindNewTimeDivisionValue(before, after, gUmt.UIParams.songTab[tab].voice[vcnum].minNote);
                    gUmt.UIParams.songTab[tab].voice[vcnum].maxNote = umtUiFindNewTimeDivisionValue(before, after, gUmt.UIParams.songTab[tab].voice[vcnum].maxNote);
                    gUmt.UIParams.songTab[tab].voice[vcnum].chunkSize = umtUiFindNewTimeDivisionValue(before, after, gUmt.UIParams.songTab[tab].voice[vcnum].chunkSize);
                    gUmt.UIParams.songTab[tab].voice[vcnum].frameSize = umtUiFindNewTimeDivisionValue(before, after, gUmt.UIParams.songTab[tab].voice[vcnum].frameSize);
                    // This section copies the symmetry pattern sections as a default setting, but doesn't override if already existing
                    for (bidx in before) {
                        if (before.hasOwnProperty(bidx)) {
                            mapFrom = before[bidx];
                            if (gUmt.UIParams.songTab[tab].voice[vcnum].symmetry.hasOwnProperty(mapFrom)) {
                                if (after.hasOwnProperty(bidx)) {
                                    mapTo = after[bidx];
                                    if (!gUmt.UIParams.songTab[tab].voice[vcnum].symmetry.hasOwnProperty(mapTo)) {
                                        umtUiCopySymmetry(tab, vcnum, mapFrom, tab, vcnum, mapTo);
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}

function umtUiDeleteVoice(voiceNum) {
    "use strict";
    var tab, vcnum;
    gUmt.UIParams.parts = gUmt.UIParams.parts - 1;
    for (tab = 0; tab < gUmt.UIParams.numTabs; tab = tab + 1) {
        for (vcnum = voiceNum; vcnum < gUmt.UIParams.parts; vcnum = vcnum + 1) {
            umtUiCopyTabVoice(tab, vcnum + 1, tab, vcnum);
        }
    }
    if (gUmt.noReenterAddingVoice === false) {
        umtAutomaticallyRecomposeLoopFromUIParams();
    }
    umtUiCopyUiParamsToActualUi();
}

umtExecDeleteVoiceClick = function () {
    "use strict";
    var elemid, UIVoiceNum;
    elemid = event.target.id;
    UIVoiceNum = Number(elemid.substring(11));
    umtUiDeleteVoice(UIVoiceNum);
};

function umtUiSetDefaultsForSymmetryForTabAndVoice(tab, voiceNum) {
    "use strict";
    var mult, keepGoing, tdvidx, currentTimeDiv, newPattern;
    mult = 1;
    keepGoing = true;
    for (tdvidx in gUmt.UIParams.timeDivisions) {
        if (gUmt.UIParams.timeDivisions.hasOwnProperty(tdvidx)) {
            currentTimeDiv = gUmt.UIParams.timeDivisions[tdvidx];
            if (currentTimeDiv === 1) {
                keepGoing = false;
            }
            if (keepGoing) {
                newPattern = {
                    pattern: "trans",
                    transParam: 0,
                    scaleParam: 0,
                    tiltParam: 0,
                    scrollrhythmParam: 0,
                    scrollpitchParam: 0,
                    scrollbothParam: 0
                };
                if (!gUmt.UIParams.songTab[tab].voice[voiceNum].symmetry.hasOwnProperty(mult)) {
                    gUmt.UIParams.songTab[tab].voice[voiceNum].symmetry[mult] = newPattern;
                }
                mult = mult * currentTimeDiv;
            }
        }
    }
}

function umtUiSetDefaultsForSymmetryForAllTabsVoices() {
    "use strict";
    var tab, vcnum;
    for (tab = 0; tab < gUmt.UIParams.numTabs; tab = tab + 1) {
        for (vcnum = 0; vcnum < gUmt.UIParams.parts; vcnum = vcnum + 1) {
            umtUiSetDefaultsForSymmetryForTabAndVoice(tab, vcnum);
        }
    }
}

umtExecTimeDivisionChange = function () {
    "use strict";
    var elemid, tdvidx, listbox, choice;
    elemid = event.target.id;
    tdvidx = Number(elemid.substring(17));
    listbox = document.getElementById("lb_time_division_" + tdvidx);
    choice = Number(listbox.options[listbox.selectedIndex].value);
    umtUiRecalculateAllTimeDivisions(tdvidx, choice);
    umtUiSetDefaultsForSymmetryForAllTabsVoices();
    if (gUmt.noReenterAddingVoice) {
        jsabort("gUmt.noReenterAddingVoice is already set to true (point 5097)");
    }
    gUmt.noReenterAddingVoice = true;
    umtUiInsertAllSymmetrySliders();
    umtUiAttachSymmetryEventListeners();
    umtUiCopyUiParamsToActualUi(); // this call will clear gUmt.noReenterAddingVoice
    gUmt.noReenterAddingVoice = false; // redundant because umtUiCopyUiParamsToActualUi clears it
    umtSetAllTabsAllLoopsToRecompose();
    umtAutomaticallyRecomposeLoopFromUIParams();
};

function umtExecAddVoice() {
    "use strict";
    var tab, voiceNum, songnum, dup;
    tab = gUmt.UIParams.currentTab;
    voiceNum = gUmt.UIParams.parts;
    songnum = voiceNum + 1;
    if (voiceNum === 0) {
        gUmt.UIParams.songTab[tab].voice[voiceNum] = {
            playing: false,
            exemptFromHarmonization: false,
            instrument: "tuningfork",
            volume: 1.0,
            songNumber: songnum,
            octave: 5,
            noteDistance: 0.3,
            noRepeatNotes: false,
            relativeNotes: false,
            restyness: 0.0,
            ampVariation: 0.5,
            skew: 0.0,
            minNote: 4,
            maxNote: 4,
            chunkSize: 4,
            frameSize: 16,
            symmetry: {},
            instSpecificParams: {},
            instFixedParams: {},
            copyChangesToOtherTabs: true
        };
        umtUiSetDefaultsForSymmetryForTabAndVoice(tab, voiceNum);
    } else {
        for (dup = 0; dup < gUmt.UIParams.numTabs; dup = dup + 1) {
            umtUiCopyTabVoice(tab, voiceNum - 1, dup, voiceNum);
            gUmt.UIParams.songTab[dup].voice[voiceNum].playing = false;
            gUmt.UIParams.songTab[dup].voice[voiceNum].exemptFromHarmonization = false;
            gUmt.UIParams.songTab[dup].voice[voiceNum].songNumber = songnum;
            if (dup === tab) {
                gUmt.UIParams.songTab[dup].voice[voiceNum].copyChangesToOtherTabs = true;
            }
        }
    }
    umtUiSetDefaultInstSpecificParamsForNewInstrument(voiceNum, gUmt.UIParams.songTab[tab].voice[voiceNum].instrument);
    umtUiSetDefaultInstFixedParamsForNewInstrument(voiceNum, gUmt.UIParams.songTab[tab].voice[voiceNum].instrument);
    gUmt.UIParams.parts = gUmt.UIParams.parts + 1;
    umtUiCopyUiParamsToActualUi();
}

// have to use var syntax for this function to get the forward reference to pass jslint

umtExecReSeed = function () {
    "use strict";
    var tab, currentTime, timeInMs, rng, i;
    tab = gUmt.UIParams.currentTab;
    currentTime = new Date(); // year, month [, day, hour, minute, second, millisecond]);
    timeInMs = currentTime.getTime();
    gUmt.lastReSeedTime = timeInMs;
    timeInMs = timeInMs - (Math.floor(timeInMs / 256) * 256);
    rng = umtGetRando(timeInMs);
    for (i = 0; i < gUmt.UIParams.parts; i = i + 1) {
        gUmt.UIParams.songTab[tab].voice[i].songNumber = Math.floor(rng.genrandReal2() * 50 + 1);
    }
    umtUiCopyUiParamsToActualUi();
    if (gUmt.noReenterAddingVoice === false) {
        umtAutomaticallyRecomposeLoopFromUIParams();
    }
};

function umtUiCopyTabAll(fromTab, toTab) {
    "use strict";
    var vcnum;
    gUmt.UIParams.songTab[toTab] = {
        tempo: gUmt.UIParams.songTab[fromTab].tempo,
        centernote: gUmt.UIParams.songTab[fromTab].centernote,
        chord: gUmt.UIParams.songTab[fromTab].chord,
        scale: gUmt.UIParams.songTab[fromTab].scale,
        voice: [ ]
    };
    for (vcnum = 0; vcnum < gUmt.UIParams.parts; vcnum = vcnum + 1) {
        umtUiCopyTabVoice(fromTab, vcnum, toTab, vcnum);
    }
}

function umtUiClearAllCopyFlags() {
    "use strict";
    var tab, vcnum;
    for (tab = 0; tab < gUmt.UIParams.numTabs; tab = tab + 1) {
        for (vcnum = 0; vcnum < gUmt.UIParams.parts; vcnum = vcnum + 1) {
            gUmt.UIParams.songTab[tab].voice[vcnum].copyChangesToOtherTabs = false;
        }
    }
}

function umtUiAddTab() {
    "use strict";
    umtUiClearAllCopyFlags();
    umtUiCopyTabAll(gUmt.UIParams.currentTab, gUmt.UIParams.numTabs);
    gUmt.UIParams.numTabs = gUmt.UIParams.numTabs + 1;
    if (gUmt.UIParams.playAllTabs) {
        umtComposeTab(0, gUmt.UIParams.numTabs - 1);
        umtComposeTab(1, gUmt.UIParams.numTabs - 1);
    }
}

function umtUiDeleteTab(tab) {
    "use strict";
    var i;
    gUmt.UIParams.numTabs = gUmt.UIParams.numTabs - 1;
    for (i = tab; i < gUmt.UIParams.numTabs; i = i + 1) {
        umtUiCopyTabAll(i + 1, i);
    }
}

function umtUiHiddenCopyBetweenTabs() {
    "use strict";
    var tab, vcnum, inner;
    // This function assumes there cannot be more than one "copyChangesToOtherTabs" set on any given voice for all tabs
    for (tab = 0; tab < gUmt.UIParams.numTabs; tab = tab + 1) {
        for (vcnum = 0; vcnum < gUmt.UIParams.parts; vcnum = vcnum + 1) {
            if (gUmt.UIParams.songTab[tab].voice[vcnum].copyChangesToOtherTabs) {
                for (inner = 0; inner < gUmt.UIParams.numTabs; inner = inner + 1) {
                    if (inner !== tab) {
                        umtUiCopyTabVoice(tab, vcnum, inner, vcnum);
                    }
                }
            }
        }
    }
    umtUiClearAllCopyFlags();
}

function umtExecLoopTabsOne() {
    "use strict";
    gUmt.UIParams.playAllTabs = false;
    umtPlaybackHighlightOneTab(-1);
    umtAutomaticallyRecomposeLoopFromUIParams();
}

function umtExecLoopTabsAll() {
    "use strict";
    gUmt.UIParams.playAllTabs = true;
    umtAutomaticallyRecomposeLoopFromUIParams();
}

function umtCloseSaveAsMessage() {
    "use strict";
    jQuery("#savedia").dialog("close");
}

function umtExecSaveAs() {
    "use strict";
    var uiparams, filename;
    uiparams = JSON.stringify(gUmt.UIParams);
    filename = document.getElementById("save_name").value;
    if (filename === "") {
        return;
    }
    jQuery(function () {
        jQuery("#savedia").dialog();
    });
    jQuery.ajax("saveuiparams", {
        method: "POST",
        data: { name: filename, uiparams: uiparams },
        dataType: "json",
        error: function (jqXHR, textStatus, errorThrown) {
            console.log("jqXHR", jqXHR);
            console.log("textStatus", textStatus);
            console.log("errorThrown", errorThrown);
            document.getElementById("saveresult").innerHTML = errorThrown;
        },
        success: function (data, textStatus, jqXHR) {
            console.log("data", data);
            console.log("textStatus", textStatus);
            console.log("jqXHR", jqXHR);
            if (!data.success) {
                document.getElementById("saveresult").innerHTML = data.error;
                return;
            }
            document.getElementById("saveresult").innerHTML = "Saved.";
            window.setTimeout(umtCloseSaveAsMessage, 800);
        }
    });
}

function umtConnectWebSocket() {
    "use strict";
    gUmt.localSocket = new WebSocket("ws://127.0.0.1:46398/umtlocal");
    gUmt.localSocket.onmessage = function (event) {
        console.log(event.data);
    };
}
function umtExecTestReconnect() {
    "use strict";
    umtConnectWebSocket();
}

var tabContent, tabCounter, tabTemplate, tabs;

// tabs
// I tried for hours to get the code to work with just one addTab() function -- for both the user pushing the "Add Tab" button, and us here adding tabs when loading a paramset
// but I gave up and just made a duplicate of the function.
// well, duplicate but without umtUiAddTab();
function addTab2() {
    "use strict";
    // var label, id, li, tabContentHtml;
    // label = tabTitle.val() || "Tab " + tabCounter;
    var label, id, li, tabContentHtml, nextTabNum, nextTabStr, i, c;
    nextTabNum = tabCounter - 1; // gUmt.UIParams.numTabs;
    nextTabStr = nextTabNum.toString(26);
    label = "";
    for (i = 0; i < nextTabStr.length; i = i + 1) {
        c = nextTabStr.charCodeAt(i);
        if (c > 57) {
            label = label + String.fromCharCode(c - 22);
        } else {
            label = label + String.fromCharCode(c + 17);
        }
    }
    id = "tabs-" + tabCounter;
    li = jQuery(tabTemplate.replace(/#\{href\}/g, "#" + id).replace(/#\{label\}/g, label));
    tabContentHtml = tabContent.val() || "Tab " + tabCounter + " content.";
    tabContentHtml = "";
    tabs.find(".ui-tabs-nav").append(li);
    tabs.append("<div id='" + id + "'><p>" + tabContentHtml + "</p></div>");
    tabs.tabs("refresh");
    tabCounter = tabCounter + 1;
}

function umtLoadParamset(paramsetid) {
    "use strict";
    jQuery.ajax("loaduiparams", {
        method: "GET",
        data: { paramset: paramsetid },
        dataType: "json",
        error: function (jqXHR, textStatus, errorThrown) {
            jQuery("#loaddia").dialog();
            document.getElementById("loadresult").innerHTML = errorThrown;
            console.log(jqXHR);
            console.log(textStatus);
        },
        success: function (data, textStatus, jqXHR) {
            var tab, numTabs;
            if (!data.success) {
                jQuery("#loaddia").dialog();
                document.getElementById("loadresult").innerHTML = data.error;
                return;
            }
            console.log("data", data);
            console.log("textStatus", textStatus);
            console.log("jqXHR", jqXHR);
            gUmt.UIParams = data.uiparams;
            numTabs = gUmt.UIParams.numTabs;
            cx("numTabs = " + ctstr(numTabs));
            for (tab = 1; tab < numTabs; tab = tab + 1) {
                cx("tab " + ctstr(tab));
                addTab2();
            }
            umtUiCopyUiParamsToActualUi();
        }
    });
}

// main

// calibration and centernote need to be the same
gUmt = {
    TAU: Math.PI * 2,
    LOG2: Math.log(2),
    calibration: 440,
    globalCtx: umtGetAudioContext(),
    globalRng: umtGetRando(0),
    cachedNotes: {},
    cachedWads: {},
    instrumentBank: {},
    cachedFittings: {},
    loop: [],
    playedUpTo: 0,
    timerIsOn: false,
    currentlyPlayingLoop: 0,
    compositionLag: 0,
    noReenterCompose: false,
    noReenterAddingVoice: false,
    UIParams: {
        mastervol: 0.9,
        songTab: [
            {
                tempo: 360,
                scale: 'maj7reg',
                chord: "triad",
                centernote: 440,
                voice: []
            }
        ],
        parts: 0,
        currentTab: 0,
        numTabs: 1,
        playAllTabs: false,
        timeDivisions: [ 2, 2, 2, 2, 2, 2 ]
    },
    tabRecomposeFlags: [],
    localSocket: false
};
// for fixing duplicate notes bug
// testDupDetectSet: {}

// console.log(gUmt);

jQuery(function () {
    "use strict";
    // var tabContent, tabTemplate, tabCounter, tabs, dialog, form;
    var dialog, form;
    jQuery("#slider_tempo").slider({min: 60, max: 960, step: 30, value: 360, orientation: "horizontal", change: function (event, ui) {
        var tab, newvalue;
        tab = gUmt.UIParams.currentTab;
        if (gUmt.noReenterAddingVoice === false) {
            if (event.type === "slidechange") {
                newvalue = ui.value;
                document.getElementById("display_tempo_value").innerHTML = newvalue;
                gUmt.UIParams.songTab[tab].tempo = newvalue;
                umtSetAllTabsAllLoopsToRecompose();
                umtAutomaticallyRecomposeLoopFromUIParams();
            }
        }
    }
         });
    jQuery("#slider_centernote").slider({min: 1, max: 9, step: 1, value: 5, orientation: "horizontal", change: function (event, ui) {
        var tab, newvalue;
        tab = gUmt.UIParams.currentTab;
        if (gUmt.noReenterAddingVoice === false) {
            if (event.type === "slidechange") {
                newvalue = ui.value;
                newvalue = Math.pow((5 / 4), newvalue - 5) * gUmt.calibration;
                document.getElementById("display_centernote_value").innerHTML = newvalue;
                gUmt.UIParams.songTab[tab].centernote = newvalue;
                umtAutomaticallyRecomposeLoopFromUIParams();
            }
        }
    }
         });
    jQuery("#slider_mastervol").slider({min: 0, max: 100, step: 5, value: 90, orientation: "horizontal", change: function (event, ui) {
        var newvalue;
        if (gUmt.noReenterAddingVoice === false) {
            if (event.type === "slidechange") {
                newvalue = ui.value;
                document.getElementById("display_mastervol_value").innerHTML = newvalue;
                gUmt.UIParams.mastervol = newvalue / 100.0;
                umtSetAllTabsAllLoopsToRecompose();
                umtAutomaticallyRecomposeLoopFromUIParams();
            }
        }
    }
         });

    // tabs
    // actual addTab function: adds new tab using the input from the form above
    function addTab() {
        // var label, id, li, tabContentHtml;
        // label = tabTitle.val() || "Tab " + tabCounter;
        var label, id, li, tabContentHtml, nextTabNum, nextTabStr, i, c;
        nextTabNum = tabCounter - 1; // gUmt.UIParams.numTabs;
        nextTabStr = nextTabNum.toString(26);
        label = "";
        for (i = 0; i < nextTabStr.length; i = i + 1) {
            c = nextTabStr.charCodeAt(i);
            if (c > 57) {
                label = label + String.fromCharCode(c - 22);
            } else {
                label = label + String.fromCharCode(c + 17);
            }
        }
        id = "tabs-" + tabCounter;
        li = jQuery(tabTemplate.replace(/#\{href\}/g, "#" + id).replace(/#\{label\}/g, label));
        tabContentHtml = tabContent.val() || "Tab " + tabCounter + " content.";
        tabContentHtml = "";
        tabs.find(".ui-tabs-nav").append(li);
        tabs.append("<div id='" + id + "'><p>" + tabContentHtml + "</p></div>");
        tabs.tabs("refresh");
        tabCounter = tabCounter + 1;
        umtUiAddTab();
    }

    // tabTitle = jQuery("#tab_title");
    tabContent = jQuery("#tab_content");
    tabTemplate = "<li><a href='#{href}'>#{label}</a> <span class='ui-icon ui-icon-close' role='presentation'>Remove Tab</span></li>";
    tabCounter = 2;

    tabs = jQuery("#tabs").tabs({
        activate: function (event, ui) {
            var active;
            console.log(event);
            console.log(ui);
            active = jQuery("#tabs").tabs("option", "active");
            // window.alert(active);
            gUmt.UIParams.currentTab = active;
            umtUiHiddenCopyBetweenTabs();
            umtUiCopyUiParamsToActualUi();
            if (!gUmt.UIParams.playAllTabs) {
                umtAutomaticallyRecomposeLoopFromUIParams();
            }
        }
    });

    // modal dialog init: custom buttons and a "close" callback resetting the form inside
    dialog = jQuery("#tabdia").dialog({
        autoOpen: false,
        modal: true,
        buttons: {
            Add: function () {
                addTab();
                jQuery(this).dialog("close");
            },
            Cancel: function () {
                jQuery(this).dialog("close");
            }
        },
        close: function () {
            form[0].reset();
        }
    });

    // addTab form: calls addTab function on submit and closes the dialog
    form = dialog.find("form").submit(function (event) {
        addTab();
        dialog.dialog("close");
        event.preventDefault();
    });

    // addTab button: just opens the dialog
    jQuery("#add_tab")
        .button()
        .click(function () {
            // dialog.dialog("open");
            addTab();
        });

    // close icon: removing the tab on click
    tabs.delegate("span.ui-icon-close", "click", function () {
        var panelId, kids, count, tabNumber;
        panelId = jQuery(this).closest("li").remove().attr("aria-controls");
        count = 0;
        tabNumber = -1;
        kids = tabs.children();
        kids.each(function (index) {
            // console.log(index + ": " + jQuery(this).id);
            console.log(index + ": " + this.id);
            if (this.id === panelId) {
                // window.alert("found!!");
                tabNumber = count;
            }
            if (this.id.substring(0, 5) === "tabs-") {
                count = count + 1;
            }
        });
        // window.alert("Tab removed: " + ctstr(tabNumber));
        umtUiDeleteTab(tabNumber);
        umtAutomaticallyRecomposeLoopFromUIParams();
        jQuery("#" + panelId).remove();
        tabs.tabs("refresh");
    });

    tabs.bind("keyup", function (event) {
        if (event.altKey && event.keyCode === jQuery.ui.keyCode.BACKSPACE) {
            var panelId = tabs.find(".ui-tabs-active").remove().attr("aria-controls");
            jQuery("#" + panelId).remove();
            tabs.tabs("refresh");
        }
    });

    umtConnectWebSocket();

    // start song
    umtCreateInstrumentBank();
`)
	if paramsetid == 0 {
		fmt.Fprintln(w, `
    umtExecAddVoice();
    document.getElementById("playvoice0").checked = true;
    gUmt.UIParams.songTab[0].voice[0].playing = true;
`)
	} else {
		fmt.Fprintln(w, `
    umtLoadParamset(`+strconv.FormatUint(paramsetid, 10)+`);
`)
	}
	fmt.Fprintln(w, `
    // umtExecAutomaticStart();
});
`)
	fmt.Fprintln(w, `

</script>

</head>
<body>
<div id="savedia" title="Saving...">
	<p id="saveresult"></p>
</div>
<div id="loaddia" title="Load Error">
	<p id="loadresult"></p>
</div>
<div id="tabdia" title="Tab data">
	<form>
		<fieldset class="ui-helper-reset">
			<label for="tab_title">Title</label>
			<input type="text" name="tab_title" id="tab_title" value="Tab Title" class="ui-widget-content ui-corner-all">
			<label for="tab_content">Content</label>
			<textarea name="tab_content" id="tab_content" class="ui-widget-content ui-corner-all">Tab content</textarea>
		</fieldset>
	</form>
</div>

  <section>
    <h1>Ultimate Music Toy</h1>

<p>
    <input type="button" id="start_automatic" value="Start Music" />
    <input type="button" id="stop_automatic" value="Stop Music" />
    <!-- input type="button" id="test_play_raw" value="Test Play Raw" -->
    <!-- input type="button" id="test_dump_gumt" value="Test Dump gUmt" -->
    <!-- input type="button" id="test_run_tests" value="Run Tests" -->
    <!-- input type="button" id="test_symmetry_system" value="Test Symmetry System" -->
    Composition time (RA) <input type="text" id="test_composition_time" value="0" />

    &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Save as: <input type="text" id="save_name" name="save_name" value="`)
	if userid == 0 {
		fmt.Fprintln(w, "You cannot save because you are not logged in")
	} else {
		fmt.Fprintln(w, name)
	}
	fmt.Fprintln(w, `" /> <input type="button" id="save_as" value="Save As" />
&nbsp;&nbsp;&nbsp;&nbsp;
<input type="button" id="test_reconnect" value="Reconnect" />

</p>

<hr />

<p> Master volume control:

<div id="slider_mastervol" class="ddrg" style="width:200px;"></div>

<div id="display_mastervol_value">90</div>

</p>

<p> <input type="radio" name="loop_tabs" value="one" id="loop_tabs_one" checked="checked" /> Play selected tab <input type="radio" name="loop_tabs" value="all" id="loop_tabs_all" /> Play all tabs </p>

<button id="add_tab">Add Tab</button>

<!-- select id="lb_tab_length"
    option value="1" 1 /option
    option value="2" 2 /option
    option value="4" 4 /option
    option value="8" 8 /option
    option value="16" 16 /option
    option value="32" 32 /option
    option value="64" 64 /option
    option value="128" 128 /option
    option value="256" 256 /option
    option value="512" 512 /option
    option value="1024" 1024 /option
/select -->

Tab time divisions:

<div id="time_division_list" style="display:inline;">
</div>

<div id="time_division_template" style="display:none;">
	<select id="lb_time_division_xx">
		<option value="1">stop</option>
		<option value="2">2</option>
		<option value="3">3</option>
		<option value="5">5</option>
		<option value="7">7</option>
		<option value="11">11</option>
	</select>
</div>

<div id="tabs">
	<ul>
		<li><a href="#tabs-1">A</a> <span class="ui-icon ui-icon-close" role="presentation">Remove Tab</span></li>
	</ul>


<p> Scale:

<select id="lb_scale">
    <option value="steps591">5 steps from tonic 91*</option>
    <option value="steps461">4 steps from tonic 61*</option>
    <option value="steps337">3 steps from tonic 37*</option>
    <option value="steps219">2 steps from tonic 19*</option>
    <option value="chroma15tt">Chromatic 15 (alt)*</option>
    <option value="chroma13">Chromatic 13 (alt)</option>
    <option value="chroma12regregtt">Chromatic 12*</option>
    <option value="chroma11regreg">Chromatic 11</option>
    <option value="hungarian7tt">Hungarian Minor 7*</option>
    <option value="arabic7">Arabic 7</option>
    <option value="freygish7">Freygish Phrygian Dominant 7</option>
    <option value="ukranian7">Ukranian Dorian 7</option>
    <option value="minharm7reg">Minor Harmonic 7</option>
    <option value="minnat7regreg">Minor Natural 7</option>
    <option value="maj7reg" selected="selected">Major 7</option>
    <option value="steps17">1 step from tonic 7</option>
    <option value="ionian7">Ionian Mode 7</option>
    <option value="dorian7">Dorian Mode 7</option>
    <option value="phrygian7">Phrygian Mode 7</option>
    <option value="lydian7tt">Lydian Mode 7*</option>
    <option value="mixolydian7">Mixolydian Mode 7</option>
    <option value="aeolian7">Aeolian Mode 7</option>
    <option value="locrian7tt">Locrian Mode 7*</option>
    <option value="pentatonic23reg">Pentatonic 5 (2+3 standard)</option>
    <option value="pentatonic131reg">Pentatonic 5 (1+3+1)</option>
    <option value="pentatonic32reg">Pentatonic 5 (3+2)</option>
    <option value="pentatonic221reg">Pentatonic 5 (2+2+1)</option>
    <option value="pentatonic122reg">Pentatonic 5 (1+2+2)</option>
    <option value="quad">Quadtonic 4</option>
    <option value="tri">Tritonic 3</option>
    <option value="tritone2test">Tritone Test 2*</option>
</select>

* = contains tritone

</p>

<p> Harmonization:

<select id="lb_harmonization">
    <option value="perfect">Perfect</option>
    <option value="major">Major</option>
    <option value="triad" selected="selected">Triads</option>
    <option value="reg">7th chords</option>
    <option value="alt">More 7th chords</option>
    <option value="any">Allow dissonance</option>
</select>

</p>

<p> Center note (tonic/key):

<div id="slider_centernote" class="ddrg" style="width:200px;"></div>

<div id="display_centernote_value">X</div>

</p>

<p>

<p> Tempo:

<div id="slider_tempo" class="ddrg" style="width:200px;"></div>

<div id="display_tempo_value">360</div>

</p>

<p>

<input type="button" id="reseed" value="Re-randomize" />

<input type="checkbox" id="continuous" /> Continuous

</p>

<hr />

<div id="voice_template" style="display:none;" >
	<p><b>This is the template row</b></p>
	<table border="0">
		<tr>
			<td id="voicedescriptcellxx" valign="top">
				Voice #XX
			</td><td id="voiceplaycellxx" valign="top">
				Play
				<br />
				<input type="checkbox" name="playvoicexx" id="playvoicexx">
			</td><td id="voiceinstrumentcellxx" valign="top" align="center">
				Instrument: <br />
				<select name="instrumentxx" id="instrumentxx">
					<option value="tuningfork">Sine wave</option>
					<option value="sinestartstop">Sine start/stop </option>
					<!-- option value="squarewave" Square /option -->
					<option value="squarevar">Square wave</option>
					<!-- option value="trianglewave" Triangle /option -->
					<option value="squarerisingpitch">Square w/rising pitch</option>
					<option value="squarestartstop">Square Start/Stop</option>
					<option value="trianglevar">Triangle wave</option>
					<option value="trianglestartstop">Triangle start/stop</option>
					<!-- option value="sawtoothwave" Sawtooth /option -->
					<option value="sawtoothvar">Sawtooth wave</option>
					<option value="sawtoothstartstop">Sawtooth start/stop</option>
					<!-- option value="forwardsnoise" Noise /option -->
					<!-- option value="backwardsnoise" Backwards noise /option -->
					<option value="steampop">Slapping pencil</option>
					<!-- option value="fmsynthoriginal" FM Synthesis /option -->
					<option value="fmsynthvar">FM Synthesis</option>
					<option value="fmsynthstartstop">FM Synthesis/Start/Stop</option>
					<option value="fmsynthsquare">FM Synthesis Square</option>
					<option value="fmsynthcrossnote">FM Synthesis/Cross note</option>
					<!-- option value="wadsin" WAD Sine /option -->
					<!-- option value="wadkick" WAD Kick /option -->
					<option value="wadbass">WAD Bass</option>
					<option value="wadsnare">WAD Snare</option>
					<option value="wadhihatopen">WAD Hihat Open</option>
					<option value="wadhihatclosed">WAD Hihat Closed</option>
					<option value="wadflute">WAD Flute</option>
					<option value="wadpiano">WAD Piano</option>
					<option value="danlights">Dan Lights</option>
					<option value="fadecandy">Fadecandy</option>
				</select>
				<div id="fixedparamsxx"></div>
			</td><td id="voiceexemptharmonizationcellxx" valign="top">
				EFH
				<br />
				<input type="checkbox" name="voiceexemptfromharmonizatonxx" id="voiceexemptfromharmonizationxx">
			</td><td id="voicevolumecellxx" valign="top" align="center">
				Volume: <br />
				<div id="slider_volumexx" class="ddrg" style="height:100px;"></div> <br />
				<div id="display_volumexx"></div>
			</td><td id="voicesongnumcellxx" valign="top" align="center">
				Seed: <br />
				<select name="songnumberxx" id="songnumberxx">
					<option value="1">1</option>
					<option value="2">2</option>
					<option value="3">3</option>
					<option value="4">4</option>
					<option value="5">5</option>
					<option value="6">6</option>
					<option value="7">7</option>
					<option value="8">8</option>
					<option value="9">9</option>
					<option value="10">10</option>
					<option value="11">11</option>
					<option value="12">12</option>
					<option value="13">13</option>
					<option value="14">14</option>
					<option value="15">15</option>
					<option value="16">16</option>
					<option value="17">17</option>
					<option value="18">18</option>
					<option value="19">19</option>
					<option value="20">20</option>
					<option value="21">21</option>
					<option value="22">22</option>
					<option value="23">23</option>
					<option value="24">24</option>
					<option value="25">25</option>
					<option value="26">26</option>
					<option value="27">27</option>
					<option value="28">28</option>
					<option value="29">29</option>
					<option value="30">30</option>
					<option value="31">31</option>
					<option value="32">32</option>
					<option value="33">33</option>
					<option value="34">34</option>
					<option value="35">35</option>
					<option value="36">36</option>
					<option value="37">37</option>
					<option value="38">38</option>
					<option value="39">39</option>
					<option value="40">40</option>
					<option value="41">41</option>
					<option value="42">42</option>
					<option value="43">43</option>
					<option value="44">44</option>
					<option value="45">45</option>
					<option value="46">46</option>
					<option value="47">47</option>
					<option value="48">48</option>
					<option value="49">49</option>
					<option value="50">50</option>
				</select>
			</td><td id="voiceoctavecellxx" valign="top" align="center">
				Octave: <br />
				<div id="slider_octavexx" class="ddrg" style="height:100px;"></div> <br />
				<div id="display_octavexx"></div>
			</td><td id="voicenotedistcellxx" valign="top" align="center">
				Note Distance:<br />
				<div id="slider_notedistancexx" class="ddrg" style="height:100px;"></div> <br />
				<div id="display_notedistancexx"></div>
			</td><td id="voicenorepeatcellxx" valign="top" align="center">
				NR:<br />
				<input type="checkbox" name="voicenorepeatnotesxx" id="voicenorepeatnotesxx">
				<br />Rel:<br />
				<input type="checkbox" name="voicerelativenotesxx" id="voicerelativenotesxx">
			</td><td id="voicerestynesscellxx" valign="top" align="center">
				Restyness:<br />
				<div id="slider_restynessxx" class="ddrg" style="height:100px;"></div> <br />
				<div id="display_restynessxx"></div>
			</td><td id="voiceampvariationcellxx" valign="top" align="center">
				Ampl. Var.:<br />
				<div id="slider_ampvariationxx" class="ddrg" style="height:100px;"></div> <br />
				<div id="display_ampvariationxx"></div>
			</td><td id="voiceskewcellxx" valign="top" align="center">
				Skew:<br />
				<div id="slider_skewxx" class="ddrg" style="height:100px;"></div> <br />
				<div id="display_skewxx"></div>
			</td><td id="voiceminnotecellxx" valign="top" align="center">
				Min Note:<br />
				<div id="slider_minnotexx" class="ddrg" style="height:100px;"></div> <br />
				<div id="display_minnotexx"></div>
			</td><td id="voicemaxnotecellxx" valign="top" align="center">
				Max Note:<br />
				<div id="slider_maxnotexx" class="ddrg" style="height:100px;"></div> <br />
				<div id="display_maxnotexx"></div>
			</td><td id="voicechunksizecellxx" valign="top" align="center">
				Chunk Size:<br />
				<div id="slider_chunksizexx" class="ddrg" style="height:100px;"></div> <br />
				<div id="display_chunksizexx"></div>
			</td><td id="voiceframesizecellxx" valign="top" align="center">
				Pattern Size:<br />
				<div id="slider_framesizexx" class="ddrg" style="height:100px;"></div> <br />
				<div id="display_framesizexx"></div>
			</td><td id="voicesymmetriescellxx" valign="top" align="center">
				Variations: <br />
			</td><td id="instparameterscellxx" valign="top" align="center">
				<!-- inst_specific_param_ins_point -->
			</td>
			<td id="voicedeletecellxx" valign="top" align="center">
				<input type="button" name="deletevoicexx" id="deletevoicexx" value="X">
			</td>
		</tr>
	</table>
</div>

<div id="symmetry_variations_template" style="display:none;">
	<table border="0" id="tbl_symmetryxx_yyy">
		<tr>
			<td valign="center" align="right">
				<span id="display_symmetrylevelxx_yyy" style="display:none;">1</span>
			</td><td valign="top">
				<select name="sel_symmetryxx_yyy" id="sel_symmetryxx_yyy" style="display:none;">
					<option value="trans">Transpose</option>
					<option value="invert">Invert</option>
					<option value="scale">Scale</option>
					<option value="tilt">Tilt</option>
					<option value="reverseboth">Reverse</option>
					<option value="reverserhythm">Reverse-rhythm</option>
					<option value="reversepitch">Reverse-pitch</option>
					<option value="scrollboth">Scroll</option>
					<option value="scrollrhythm">Scroll-rhythm</option>
					<option value="scrollpitch">Scroll-pitch</option>
				</select>
			</td><td valign="top" align="center">
				<div id="slider_symmetry_transParamxx_yyy" class="ddrg" style="width:100px;display:none;"></div>
				<span id="display_symmetry_transParamxx_yyy" style="display:none;">readout</span>
				<div id="slider_symmetry_scaleParamxx_yyy" class="ddrg" style="width:100px;display:none;"></div>
				<span id="display_symmetry_scaleParamxx_yyy" style="display:none;">readout</span>
				<div id="slider_symmetry_tiltParamxx_yyy" class="ddrg" style="width:100px;display:none;"></div>
				<span id="display_symmetry_tiltParamxx_yyy" style="display:none;">readout</span>
				<div id="slider_symmetry_scrollrhythmParamxx_yyy" class="ddrg" style="width:100px;display:none;"></div>
				<span id="display_symmetry_scrollrhythmParamxx_yyy" style="display:none;">readout</span>
				<div id="slider_symmetry_scrollpitchParamxx_yyy" class="ddrg" style="width:100px;display:none;"></div>
				<span id="display_symmetry_scrollpitchParamxx_yyy" style="display:none;">readout</span>
				<div id="slider_symmetry_scrollbothParamxx_yyy" class="ddrg" style="width:100px;display:none;"></div>
				<span id="display_symmetry_scrollbothParamxx_yyy" style="display:none;">readout</span>
			</td>
		</tr>
	</table>
</div>

<div id="instrument_parameter_template" style="display:none;" >
	<table border="0"><tr>
		<td>
			<div id="slider_instr_ZZ_YY_XX_patterns" class="ddrg" style="height:100px;"></div> <br />
			<div id="display_instr_ZZ_YY_XX_patterns"></div><br />
		</td>
	</tr></table>
</div>

<div id="instrument_fixed_param_template" style="display:none;" >
	<div style="display:block;">
		<select id="choose_fixed_ZZ_YY_XX">
		<!-- inst_fixed_param_listbox_options_ins_point -->
		</select>
	</div>
</div>

<div id="voices_section">
	<table id="voices_table" border="0">
		<tbody>
		</tbody>
	</table>
</div>

<input id="add_voice" type="button" value="Add a voice" />

	<div id="tabs-1"></div>
</div>

  </section>
<script>

var umtListenElement = document.getElementById("start_automatic");
umtListenElement.addEventListener("click", umtExecAutomaticStart, true);

umtListenElement = document.getElementById("stop_automatic");
umtListenElement.addEventListener("click", umtExecAutomaticStop, true);

umtListenElement = document.getElementById("lb_scale");
umtListenElement.addEventListener("change", umtExecScaleChange, true);

umtListenElement = document.getElementById("lb_harmonization");
umtListenElement.addEventListener("change", umtExecHarmonizationChange, true);

umtListenElement = document.getElementById("reseed");
umtListenElement.addEventListener("click", umtExecReSeed, true);

umtListenElement = document.getElementById("add_voice");
umtListenElement.addEventListener("click", umtExecAddVoice, true);

umtListenElement = document.getElementById("loop_tabs_one");
umtListenElement.addEventListener("click", umtExecLoopTabsOne, true);

umtListenElement = document.getElementById("loop_tabs_all");
umtListenElement.addEventListener("click", umtExecLoopTabsAll, true);

umtListenElement = document.getElementById("save_as");
umtListenElement.addEventListener("click", umtExecSaveAs, true);

// umtListenElement = document.getElementById("lb_tab_length");
// umtListenElement.addEventListener("change", umtExecLoopLengthChange, true);

// umtListenElement = document.getElementById("test_play_raw");
// umtListenElement.addEventListener("click", umtExecPlayRawScore, true);

// umtListenElement = document.getElementById("test_dump_gumt");
// umtListenElement.addEventListener("click", umtExecTestDumtGmt, true);

// umtListenElement = document.getElementById("test_run_tests");
// umtListenElement.addEventListener("click", umtExecRunTests, true);

// umtListenElement = document.getElementById("test_symmetry_system");
// umtListenElement.addEventListener("click", umtTestSetupCopyTransTestSection, true);

umtListenElement = document.getElementById("test_reconnect");
umtListenElement.addEventListener("click", umtExecTestReconnect, true);

</script>

</body>
</html>`)

}

func dequote(str string) string {
	rv := str
	i := strings.Index(rv, `"`)
	for i >= 0 {
		rv = rv[:i-1] + rv[i+1:]
		i = strings.Index(rv, `"`)
	}
	return rv
}

func saveAsAjax(w http.ResponseWriter, r *http.Request, op string, userid uint64) {
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	if userid == 0 {
		fmt.Fprintln(w, `{ "success": false, "error": "You are not logged in." }`)
		return
	}
	method := r.Method
	if method == "POST" {
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintln(w, `{ "success": false, "error": "Point 5188 `+dequote(err.Error())+`" }`)
			return
		}
		postform := r.Form
		name := postform["name"][0]
		uiparams := postform["uiparams"][0]
		db := accessdb.GetDbConnection()
		defer db.Close()
		sql := "SELECT id_paramset FROM umt_paramset WHERE (id_user = " + strconv.FormatUint(userid, 10) + ") AND (name = '" + mysql.Escape(db, name) + "');"
		res, err := db.Start(sql)
		if err != nil {
			fmt.Fprintln(w, `{ "success": false, "error": "Point 5203 `+dequote(err.Error())+`" }`)
			return
		}
		// defer res.Close();
		row, err := res.GetRow()
		if err != nil {
			fmt.Fprintln(w, `{ "success": false, "error": "Point 5209 `+dequote(err.Error())+`" }`)
			return
		}
		if row == nil {
			stmt, err := db.Prepare("INSERT INTO umt_paramset (id_user, name, uiparams, created_gmt, updated_gmt) VALUES (?, ?, ?, ?, ?);")
			if err != nil {
				fmt.Fprintln(w, `{ "success": false, "error": "Point 5215 `+dequote(err.Error())+`" }`)
				panic("Prepare failed")
			}
			// defer stmt.Close();
			var paramsetrecord struct {
				idUser     uint64
				name       string
				uiparams   string
				createdGmt uint64
				updatedGmt uint64
			}
			paramsetrecord.idUser = userid
			paramsetrecord.name = name
			paramsetrecord.uiparams = uiparams
			paramsetrecord.createdGmt = uint64(time.Now().Unix())
			paramsetrecord.updatedGmt = uint64(time.Now().Unix())
			// stmt.Bind(paramsetrecord)
			stmt.Bind(&paramsetrecord.idUser, &paramsetrecord.name, &paramsetrecord.uiparams, &paramsetrecord.createdGmt, &paramsetrecord.updatedGmt)
			_, _, err = stmt.Exec()
			if err != nil {
				fmt.Fprintln(w, `{ "success": false, "error": "Point 5235 `+dequote(err.Error())+`" }`)
				return
			}
		} else {
			paramsetid := row.Uint64(0)
			for row != nil {
				row, err = res.GetRow()
			}
			var paramsetupdate struct {
				uiparams   string
				updatedGmt uint64
				paramSetId uint64
			}
			paramsetupdate.uiparams = uiparams
			paramsetupdate.updatedGmt = uint64(time.Now().Unix())
			paramsetupdate.paramSetId = paramsetid
			stmt, err := db.Prepare("UPDATE umt_paramset SET uiparams = ?, updated_gmt = ? WHERE id_paramset = ?;")
			if err != nil {
				fmt.Fprintln(w, `{ "success": false, "error": "Point 5253 `+dequote(err.Error())+`" }`)
				return
			}
			// defer stmt.Close();
			// stmt.Bind(paramsetupdate)
			stmt.Bind(&paramsetupdate.uiparams, &paramsetupdate.updatedGmt, &paramsetupdate.paramSetId)
			_, _, err = stmt.Exec()
			if err != nil {
				fmt.Fprintln(w, `{ "success": false, "error": "Point 5260 `+dequote(err.Error())+`" }`)
				return
			}
		}
	}
	fmt.Fprintln(w, `{ "success": true }`)
}

func list(w http.ResponseWriter, r *http.Request, op string, userid uint64) {
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintln(w, `
<!DOCTYPE html>
<html>
<head>
<meta charset=utf-8 />
<title>Ultimate Music Toy Song List</title>
</head>
<body>
`)
	db := accessdb.GetDbConnection()
	defer db.Close()
	sql := "SELECT id_paramset, name FROM umt_paramset WHERE (id_user = " + strconv.FormatUint(userid, 10) + ") ORDER BY id_user, name;"
	res, err := db.Start(sql)
	if err != nil {
		fmt.Fprintln(w, err)
	}
	first := true
	row, err := res.GetRow()
	for row != nil {
		var paramSetData struct {
			paramsetid uint64
			name       string
		}
		paramSetData.paramsetid = row.Uint64(0)
		paramSetData.name = row.Str(1)
		// fmt.Fprintln(w, "id_paramset is "+strconv.FormatUint(paramSetData.paramsetid, 10)+", name "+html.EscapeString(paramSetData.name)+"<br />")
		// fmt.Fprintln(w, "id_paramset is "+strconv.FormatUint(paramSetData.paramsetid, 10)+", name "+
		if first {
			fmt.Fprintln(w, `<table cellpadding="4">`)
			first = false
		}
		fmt.Fprintln(w, "<tr><td> "+html.EscapeString(paramSetData.name)+`</td><td> <a href="umt?paramset=`+strconv.FormatUint(paramSetData.paramsetid, 10)+`">Load</a> </td></tr>`)
		row, err = res.GetRow()
	}
	if !first {
		fmt.Fprintln(w, "</table>")
	}
	fmt.Fprintln(w, `
</body>
</html>
`)
}

func loadAjax(w http.ResponseWriter, r *http.Request, op string, userid uint64) {
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintln(w, `{ "success": false, "error": "Point 5382 `+dequote(err.Error())+`" }`)
		return
	}
	getform := r.Form
	paramsetstr := getform["paramset"][0]
	paramsetid, err := strconv.ParseUint(paramsetstr, 10, 64)
	db := accessdb.GetDbConnection()
	defer db.Close()
	// sql := "SELECT uiparams FROM umt_paramset WHERE (id_paramset = " + strconv.FormatUint(paramsetid, 10) + ");"
	sql := "SELECT uiparams FROM umt_paramset WHERE (id_paramset = ?);"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Fprintln(w, `{ "success": false, "error": "Point 5397 `+dequote(err.Error())+`" }`)
		return
	}
	sel.Bind(paramsetid)
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Fprintln(w, `{ "success": false, "error": "Point 5404 `+dequote(err.Error())+`" }`)
		return
	}
	var uiparams string
	found := false
	for _, row := range rows {
		uiparams = row.Str(0)
		found = true
	}
	if found {
		fmt.Fprintln(w, `{ "success": true, "uiparams": `+uiparams+` }`)
	} else {
		fmt.Fprintln(w, `{ "success": false, "error": "Parameter set not found." }`)
	}
}

func Handler(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	switch {
	case op == "umt":
		mainPage(w, r, op, userid)
	case op == "":
		mainPage(w, r, op, userid)
	case op == "saveuiparams":
		saveAsAjax(w, r, op, userid)
	case op == "loaduiparams":
		loadAjax(w, r, op, userid)
	case op == "list":
		if userid != 0 {
			list(w, r, op, userid)
		}
	default:
		filename := "/home/ec2-user/wayneserver/staticappcontent/umt/" + op
		static.OutputStaticFileWithContentType(w, filename)
	}
}
