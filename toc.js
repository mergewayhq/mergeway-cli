// Populate the sidebar
//
// This is a script, and not included directly in the page, to control the total size of the book.
// The TOC contains an entry for each page, so if each page includes a copy of the TOC,
// the total size of the page becomes O(n**2).
class MDBookSidebarScrollbox extends HTMLElement {
    constructor() {
        super();
    }
    connectedCallback() {
        this.innerHTML = '<ol class="chapter"><li class="chapter-item expanded "><a href="index.html"><strong aria-hidden="true">1.</strong> Overview</a></li><li class="chapter-item expanded "><a href="installation.html"><strong aria-hidden="true">2.</strong> Installation</a></li><li class="chapter-item expanded "><a href="getting-started.html"><strong aria-hidden="true">3.</strong> Getting Started</a></li><li class="chapter-item expanded "><a href="concepts.html"><strong aria-hidden="true">4.</strong> Concepts</a></li><li class="chapter-item expanded "><a href="schema-spec.html"><strong aria-hidden="true">5.</strong> Schema Format</a></li><li class="chapter-item expanded "><a href="cli-reference/index.html"><strong aria-hidden="true">6.</strong> CLI Reference</a></li><li><ol class="section"><li class="chapter-item expanded "><a href="cli-reference/init.html"><strong aria-hidden="true">6.1.</strong> mw init</a></li><li class="chapter-item expanded "><a href="cli-reference/validate.html"><strong aria-hidden="true">6.2.</strong> mw validate</a></li><li class="chapter-item expanded "><a href="cli-reference/entity-list.html"><strong aria-hidden="true">6.3.</strong> mw entity list</a></li><li class="chapter-item expanded "><a href="cli-reference/entity-show.html"><strong aria-hidden="true">6.4.</strong> mw entity show</a></li><li class="chapter-item expanded "><a href="cli-reference/config-lint.html"><strong aria-hidden="true">6.5.</strong> mw config lint</a></li><li class="chapter-item expanded "><a href="cli-reference/config-export.html"><strong aria-hidden="true">6.6.</strong> mw config export</a></li><li class="chapter-item expanded "><a href="cli-reference/list.html"><strong aria-hidden="true">6.7.</strong> mw list</a></li><li class="chapter-item expanded "><a href="cli-reference/get.html"><strong aria-hidden="true">6.8.</strong> mw get</a></li><li class="chapter-item expanded "><a href="cli-reference/create.html"><strong aria-hidden="true">6.9.</strong> mw create</a></li><li class="chapter-item expanded "><a href="cli-reference/update.html"><strong aria-hidden="true">6.10.</strong> mw update</a></li><li class="chapter-item expanded "><a href="cli-reference/delete.html"><strong aria-hidden="true">6.11.</strong> mw delete</a></li><li class="chapter-item expanded "><a href="cli-reference/export.html"><strong aria-hidden="true">6.12.</strong> mw export</a></li><li class="chapter-item expanded "><a href="cli-reference/version.html"><strong aria-hidden="true">6.13.</strong> mw version</a></li></ol></li></ol>';
        // Set the current, active page, and reveal it if it's hidden
        let current_page = document.location.href.toString().split("#")[0].split("?")[0];
        if (current_page.endsWith("/")) {
            current_page += "index.html";
        }
        var links = Array.prototype.slice.call(this.querySelectorAll("a"));
        var l = links.length;
        for (var i = 0; i < l; ++i) {
            var link = links[i];
            var href = link.getAttribute("href");
            if (href && !href.startsWith("#") && !/^(?:[a-z+]+:)?\/\//.test(href)) {
                link.href = path_to_root + href;
            }
            // The "index" page is supposed to alias the first chapter in the book.
            if (link.href === current_page || (i === 0 && path_to_root === "" && current_page.endsWith("/index.html"))) {
                link.classList.add("active");
                var parent = link.parentElement;
                if (parent && parent.classList.contains("chapter-item")) {
                    parent.classList.add("expanded");
                }
                while (parent) {
                    if (parent.tagName === "LI" && parent.previousElementSibling) {
                        if (parent.previousElementSibling.classList.contains("chapter-item")) {
                            parent.previousElementSibling.classList.add("expanded");
                        }
                    }
                    parent = parent.parentElement;
                }
            }
        }
        // Track and set sidebar scroll position
        this.addEventListener('click', function(e) {
            if (e.target.tagName === 'A') {
                sessionStorage.setItem('sidebar-scroll', this.scrollTop);
            }
        }, { passive: true });
        var sidebarScrollTop = sessionStorage.getItem('sidebar-scroll');
        sessionStorage.removeItem('sidebar-scroll');
        if (sidebarScrollTop) {
            // preserve sidebar scroll position when navigating via links within sidebar
            this.scrollTop = sidebarScrollTop;
        } else {
            // scroll sidebar to current active section when navigating via "next/previous chapter" buttons
            var activeSection = document.querySelector('#sidebar .active');
            if (activeSection) {
                activeSection.scrollIntoView({ block: 'center' });
            }
        }
        // Toggle buttons
        var sidebarAnchorToggles = document.querySelectorAll('#sidebar a.toggle');
        function toggleSection(ev) {
            ev.currentTarget.parentElement.classList.toggle('expanded');
        }
        Array.from(sidebarAnchorToggles).forEach(function (el) {
            el.addEventListener('click', toggleSection);
        });
    }
}
window.customElements.define("mdbook-sidebar-scrollbox", MDBookSidebarScrollbox);
