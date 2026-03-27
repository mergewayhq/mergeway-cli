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
        this.innerHTML = '<ol class="chapter"><li class="chapter-item expanded "><a href="index.html"><strong aria-hidden="true">1.</strong> Overview</a></li><li class="chapter-item expanded "><a href="getting-started/index.html"><strong aria-hidden="true">2.</strong> Getting Started</a></li><li><ol class="section"><li class="chapter-item expanded "><a href="getting-started/installation.html"><strong aria-hidden="true">2.1.</strong> Installation</a></li><li class="chapter-item expanded "><a href="getting-started/workspace-setup.html"><strong aria-hidden="true">2.2.</strong> Workspace setup</a></li><li class="chapter-item expanded "><a href="getting-started/schema-spec.html"><strong aria-hidden="true">2.3.</strong> Schema Format</a></li></ol></li><li class="chapter-item expanded "><a href="guides/index.html"><strong aria-hidden="true">3.</strong> How-to Guides</a></li><li><ol class="section"><li class="chapter-item expanded "><a href="guides/setup-mergeway-github.html"><strong aria-hidden="true">3.1.</strong> Set Up Mergeway with GitHub</a></li><li class="chapter-item expanded "><a href="guides/setup-mergeway-pre-commit.html"><strong aria-hidden="true">3.2.</strong> Enforce Mergeway Formatting with pre-commit</a></li></ol></li><li class="chapter-item expanded "><a href="cli-reference/index.html"><strong aria-hidden="true">4.</strong> CLI Reference</a></li><li><ol class="section"><li class="chapter-item expanded "><a href="cli-reference/init.html"><strong aria-hidden="true">4.1.</strong> mergeway-cli init</a></li><li class="chapter-item expanded "><a href="cli-reference/validate.html"><strong aria-hidden="true">4.2.</strong> mergeway-cli validate</a></li><li class="chapter-item expanded "><a href="cli-reference/entity-list.html"><strong aria-hidden="true">4.3.</strong> mergeway-cli entity list</a></li><li class="chapter-item expanded "><a href="cli-reference/entity-show.html"><strong aria-hidden="true">4.4.</strong> mergeway-cli entity show</a></li><li class="chapter-item expanded "><a href="cli-reference/config-lint.html"><strong aria-hidden="true">4.5.</strong> mergeway-cli config lint</a></li><li class="chapter-item expanded "><a href="cli-reference/config-export.html"><strong aria-hidden="true">4.6.</strong> mergeway-cli config export</a></li><li class="chapter-item expanded "><a href="cli-reference/list.html"><strong aria-hidden="true">4.7.</strong> mergeway-cli list</a></li><li class="chapter-item expanded "><a href="cli-reference/get.html"><strong aria-hidden="true">4.8.</strong> mergeway-cli get</a></li><li class="chapter-item expanded "><a href="cli-reference/create.html"><strong aria-hidden="true">4.9.</strong> mergeway-cli create</a></li><li class="chapter-item expanded "><a href="cli-reference/update.html"><strong aria-hidden="true">4.10.</strong> mergeway-cli update</a></li><li class="chapter-item expanded "><a href="cli-reference/delete.html"><strong aria-hidden="true">4.11.</strong> mergeway-cli delete</a></li><li class="chapter-item expanded "><a href="cli-reference/gen-erd.html"><strong aria-hidden="true">4.12.</strong> mergeway-cli gen-erd</a></li><li class="chapter-item expanded "><a href="cli-reference/export.html"><strong aria-hidden="true">4.13.</strong> mergeway-cli export</a></li><li class="chapter-item expanded "><a href="cli-reference/version.html"><strong aria-hidden="true">4.14.</strong> mergeway-cli version</a></li></ol></li></ol>';
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
