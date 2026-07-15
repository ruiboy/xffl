// ==UserScript==
// @name         FFL Forum Capture
// @namespace    xffl
// @version      0.1.0
// @description  Capture an FFL forum round page (all team posts) and POST it to the local FFL service for in-session preview. Slice 1 of the historical import — no DB writes.
// @match        https://www.tapatalk.com/groups/ffltkf/*
// @grant        GM_xmlhttpRequest
// @connect      localhost
// @run-at       document-idle
// ==/UserScript==

// Historical FFL import — capture tool.
//
// Install in Violentmonkey / Tampermonkey / Greasemonkey. Navigate (in your real,
// logged-in browser session) to a round thread, then click the "Capture round"
// button this script injects. It reads every post on the page and POSTs it to the
// local FFL service via GM_xmlhttpRequest (which bypasses CORS + mixed-content),
// so there is no copy/paste. The FFL DataOps "Forum Capture" panel shows the parse
// preview. Navigation stays manual — the script never fetches pages itself.

(function () {
  'use strict';

  const ENDPOINT = 'http://localhost:8081/query';

  const INGEST = `
    mutation Ingest($input: IngestFFLForumPageInput!) {
      ingestFFLForumPage(input: $input) {
        season roundTitle topicId
        posts { postId author team isTeamSubmission parseError players { name position } }
      }
    }`;

  function text(el) {
    return el ? el.textContent.trim() : '';
  }

  function detectSeason() {
    // Breadcrumb crumbs include the season sub-forum, e.g. "2025".
    const crumbs = document.querySelectorAll('#nav-breadcrumbs .crumb span[itemprop="name"]');
    for (const c of crumbs) {
      const t = c.textContent.trim();
      if (/^\d{4}$/.test(t)) return t;
    }
    return '';
  }

  function detectTopicId() {
    const input = document.querySelector('input[name="topic_id"]');
    if (input && input.value) return input.value;
    const canon = document.querySelector('link[rel="canonical"]');
    const m = canon && canon.href.match(/-t(\d+)\.html/);
    return m ? m[1] : '';
  }

  function extractPage() {
    const posts = [];
    document.querySelectorAll('div[id^="p_"]').forEach((post) => {
      const content = post.querySelector('.content');
      if (!content) return; // not a real post container
      const timeEl = post.querySelector('time[datetime]');
      posts.push({
        postId: post.id.replace(/^p_/, ''),
        author: text(post.querySelector('.postprofile span[itemprop="name"]')),
        timestamp: timeEl ? timeEl.getAttribute('datetime') : '',
        html: content.innerHTML,
      });
    });
    return {
      season: detectSeason(),
      roundTitle: text(document.querySelector('.topic-title a') || document.querySelector('.topic-title')),
      topicId: detectTopicId(),
      posts,
    };
  }

  function post(input, onDone) {
    GM_xmlhttpRequest({
      method: 'POST',
      url: ENDPOINT,
      headers: { 'Content-Type': 'application/json' },
      data: JSON.stringify({ query: INGEST, variables: { input } }),
      onload: (res) => {
        try {
          const body = JSON.parse(res.responseText);
          if (body.errors) return onDone('error: ' + body.errors[0].message);
          const page = body.data.ingestFFLForumPage;
          const teams = page.posts.filter((p) => p.isTeamSubmission).map((p) => p.team);
          onDone(`✓ ${page.roundTitle || 'round'} — ${page.posts.length} posts, teams: ${teams.join(', ') || 'none'}`);
        } catch (e) {
          onDone('bad response: ' + e.message);
        }
      },
      onerror: () => onDone('request failed — is the FFL service running on :8081?'),
    });
  }

  const btn = document.createElement('button');
  btn.textContent = 'Capture round';
  btn.style.cssText =
    'position:fixed;bottom:16px;right:16px;z-index:99999;padding:10px 14px;' +
    'background:#f57f1e;color:#fff;border:none;border-radius:6px;font:600 13px sans-serif;' +
    'cursor:pointer;box-shadow:0 2px 8px rgba(0,0,0,.3)';
  btn.addEventListener('click', () => {
    const page = extractPage();
    if (page.posts.length === 0) {
      btn.textContent = 'No posts found';
      return;
    }
    const ok = confirm(
      `Capture ${page.posts.length} posts\nSeason: ${page.season || '(unknown)'}\nRound: ${page.roundTitle}`,
    );
    if (!ok) return;
    btn.textContent = 'Capturing…';
    post(page, (msg) => {
      btn.textContent = msg;
      setTimeout(() => (btn.textContent = 'Capture round'), 6000);
    });
  });
  document.body.appendChild(btn);
})();
