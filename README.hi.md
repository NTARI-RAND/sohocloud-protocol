> सामुदायिक अनुवाद (मसौदा) — NTARI नीति P2-002, Global Multilingual Broadcast। स्रोत: README.md (अंग्रेज़ी मूल, स्नैपशॉट 2026-07-29)। P2-002 §3.1 के अनुसार क्षेत्रीय अनुरक्षक (regional maintainer) की समीक्षा हेतु लंबित, मशीन-सहायता से तैयार सामुदायिक मसौदा। §2.2 के अनुसार मुख्य तकनीकी विनिर्देश अंग्रेज़ी में ही रहते हैं।
>
> अनुवाद में कोई त्रुटि दिखे तो कृपया उसे स्वयं सुधारने में मदद करें:
> रिपॉज़िटरी को फ़ॉर्क (fork) करें और अपना सुधार पुल रिक्वेस्ट (pull request)
> के रूप में भेजें — https://github.com/NTARI-RAND/sohocloud-protocol।
> अनुवाद-सुधार हमारे लिए किसी भी अन्य योगदान जितने ही मूल्यवान हैं, और उनका
> सदा स्वागत है।

# sohocloud-protocol

SoHoLINK / Cloudy नेटवर्क के लिए **सब्सट्रेट समन्वय प्रोटोकॉल** को परिभाषित
करने वाला साझा Go मॉड्यूल: नोड पहचान, क्षमता सूचीकरण (capability listing),
और जॉब नियोजन — जिसे हर कोऑर्डिनेटर और हर फ़्रंटएंड बोलता है।

## दायरा

यह मॉड्यूल केवल **समन्वय** को नियंत्रित करता है। यह JFA सदस्य अर्थव्यवस्था
**नहीं** है — न कोई एस्क्रो, न सदस्य-जारी क्रेडिट, न प्रतिष्ठा अनुबंध
(reputation covenant), न संवाद-मुद्रित रिकॉर्ड (dialog-sealed record), न
विवाद निपटान। ये सब हर फ़्रंटएंड में रहते हैं और यहाँ जान-बूझकर अनुपस्थित
हैं।

यह एक **डिपेंडेंसी लीफ़** है: यह किसी अन्य NTARI मॉड्यूल को इम्पोर्ट नहीं
करता, और Go स्टैंडर्ड लाइब्रेरी के अलावा कुछ भी नहीं। SoHoLINK (संदर्भ
कोऑर्डिनेटर) और Cloudy (एक फ़्रंटएंड) दोनों *इसे* इम्पोर्ट करते हैं; इसके
माध्यम से कुछ भी वापस इम्पोर्ट नहीं होता। यही गुण — जो गद्य से नहीं बल्कि
इम्पोर्ट ग्राफ़ से लागू होता है — किसी एक कोऑर्डिनेटर को ऐसा हब बनने से
रोकता है जिससे होकर पूरे नेटवर्क को रूट करना पड़े (खुली समस्या #7)।

## भूमिकाएँ

फ़्रंटएंड नोड-पक्ष की सतह बोलते हैं — `SubmitListing`, `Heartbeat`,
`PollJobs`, `Decline`, `ReportJob`, `Fees` — उन मशीनों की ओर से जिन्हें उनके
सदस्य योगदान करते हैं; कोऑर्डिनेटर `Coordinator` इंटरफ़ेस को लागू करते हैं
और फ़्रंटएंड्स का समन्वय (तथा फ़ेडरेशन) करते हैं। व्यक्ति कभी वायर पर प्रकट
नहीं होते: सदस्य पहचान फ़्रंटएंड का विषय है, और इस मॉड्यूल को ज्ञात एकमात्र
पहचान वर्कलोड पहचान है — NodeID, अपनी कैनोनिकल SPIFFE बाइंडिंग के साथ
`identity/` में।

## संरचना

```
version.go            protocol version (v0, unstable)
canon/                deterministic length-prefixed signing encoder
identity/             NodeID + the canonical SPIFFE binding predicate
listing/              CapabilityListing (node-signed)
liveness/             Heartbeat (node-signed)
employment/           Assignment (coordinator-signed); Decline, JobReport (node-signed)
fees/                 FeeDeclaration (coordinator-signed)
coordinator/          the Coordinator interface — the pluggable role, no algorithm
anchor/               STUB, not built — witnessed employment-claim layer (#6)
transport/httpjson/   reference transport; imported by no core package
vectors/              conformance test: regenerates and checks testdata/vectors.json
testdata/vectors.json cross-language conformance vectors (normative fixture; see SPEC.md §10)
```

## स्थिति (ईमानदार)

- **निर्मित:** संदेश प्रकार, कैनोनिकल साइनिंग बाइट्स, ed25519 साइन/वेरिफ़ाई,
  `Coordinator` इंटरफ़ेस, और `transport/httpjson` के अंतर्गत एक संदर्भ
  HTTP+JSON ट्रांसपोर्ट।
- **कन्फ़ॉर्मेंस वेक्टर:** `testdata/vectors.json` मानकीय (normative)
  क्रॉस-लैंग्वेज कन्फ़ॉर्मेंस फ़िक्स्चर है — 25 प्रिमिटिव केस तथा सभी छह
  हस्ताक्षरित संदेश प्रकार, अपने कैनोनिकल बाइट्स और ed25519 हस्ताक्षरों
  सहित। कोई बाहरी कार्यान्वयन तभी और केवल तभी कन्फ़ॉर्मेंट है जब वह समान
  इनपुट से फ़िक्स्चर को बाइट-दर-बाइट पुनरुत्पादित करे (`SPEC.md` §10)।
  `vectors/` टेस्ट वर्तमान एनकोडरों से फ़िक्स्चर को पुनर्जनित करता है और
  किसी भी विचलन (drift) पर बिल्ड विफल कर देता है।
- **संदर्भ ट्रांसपोर्ट, वायर नहीं:** `transport/httpjson` वैकल्पिक है और इसे
  कोई कोर पैकेज इम्पोर्ट नहीं करता। कोई कन्फ़ॉर्मेंट कार्यान्वयन `SPEC.md`
  में प्रलेखित कैनोनिकल बाइट्स के अनुसार `coordinator.Coordinator` को लागू
  करके इन संदेशों को किसी भी ट्रांसपोर्ट पर बोल सकता है (MAY)।
- **स्टब, निर्मित नहीं:** `anchor/` — साक्ष्य-आधारित नियोजन-दावा परत
  (witnessed employment-claim layer) (खुली समस्या #6)। स्टैंड-इन के रूप में
  चिह्नित; इसका पैकेज दस्तावेज़ देखें।

## लाइसेंस

AGPL-3.0-or-later। `SPEC.md` कॉमन्स (साझा संपदा) का हिस्सा है: इसका अस्तित्व
इसलिए है कि इसे पढ़ा जाए, पुनः कार्यान्वित किया जाए, और चुनौती दी जाए। जिस
कॉमन्स को कोई पढ़ ही न सके, वह ऐसी स्वतंत्रता है जिसका कोई उपयोग नहीं कर
सकता।

*Network Theory Applied Research Institute, Inc. — 501(c)(3) — EIN 92-3047136 — info@ntari.org*
