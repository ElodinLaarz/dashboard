// Mock HTMX
global.htmx = {
  ajax: jest.fn()
};

// Mock DOM elements and functions
let mockElements = {};
let mockUrl;

beforeEach(() => {
  // Reset mocks
  jest.clearAllMocks();
  
  // Store the original URL constructor
  const OriginalURL = global.URL;
  
  // Mock URL and URLSearchParams - need to properly track state
  mockUrl = {
    href: 'http://localhost:8080/items?groupBy=shape',
    search: '?groupBy=shape'
  };
  
  // Create a proper URL mock that updates href when search changes
  global.URL = function(urlString) {
    // Use the real URL constructor from jsdom
    const realUrl = new OriginalURL(urlString || mockUrl.href);
    
    const urlObj = {
      href: realUrl.href,
      searchParams: realUrl.searchParams,
      toString: function() { 
        return this.href;
      }
    };
    
    // When search is set, update href
    Object.defineProperty(urlObj, 'search', {
      get: function() { 
        return realUrl.search;
      },
      set: function(value) { 
        realUrl.search = value;
        urlObj.href = realUrl.href;
      },
      enumerable: true,
      configurable: true
    });
    
    return urlObj;
  };
  
  // Mock window with writable location
  delete global.window.location;
  global.window.location = new URL(mockUrl.href);
  
  // Make location properties writable
  Object.defineProperty(window, 'location', {
    value: {
      ...window.location,
      href: mockUrl.href,
      search: mockUrl.search,
      origin: 'http://localhost',
      hostname: 'localhost',
      protocol: 'http:',
      toString: () => window.location.href
    },
    writable: true
  });
  
  // Spy on history.pushState to update the URL while remaining a real History object
  jest.spyOn(window.history, 'pushState').mockImplementation((state, unused, url) => {
    if (url) {
      const newUrl = new URL(url, 'http://localhost:8080');
      window.location.href = newUrl.href;
      window.location.search = newUrl.search;
    }
  });

  // Provide no-op spies for other history methods used in tests (if any)
  if (!jest.isMockFunction(window.history.replaceState)) {
    jest.spyOn(window.history, 'replaceState').mockImplementation(() => {});
  }
  if (!jest.isMockFunction(window.history.back)) {
    jest.spyOn(window.history, 'back').mockImplementation(() => {});
  }
  if (!jest.isMockFunction(window.history.forward)) {
    jest.spyOn(window.history, 'forward').mockImplementation(() => {});
  }
  if (!jest.isMockFunction(window.history.go)) {
    jest.spyOn(window.history, 'go').mockImplementation(() => {});
  }
  
  // Mock document
  global.document = {
    getElementById: jest.fn((id) => {
      if (id === 'active-filters-container') {
        return {
          innerHTML: ''
        };
      }
      return null;
    }),
    querySelectorAll: jest.fn(() => [])
  };
  
  // Set up the setActiveFilter function
  global.setActiveFilter = function(filterType, filterValue) {
    // Get current URL parameters
    const url = new URL(window.location.href);
    const params = new URLSearchParams(url.search);
    
    // Get current filters as an array of {type, value} objects
    const currentFilters = [];
    
    // Get existing filters from URL
    for (const filter of params.getAll('filter')) {
      const [type, value] = filter.split(':');
      if (type && value) {
        currentFilters.push({ type, value });
      }
    }
    
    // Check if this exact filter already exists
    const existingFilterIndex = currentFilters.findIndex(
      f => f.type === filterType && f.value === filterValue
    );
    
    if (existingFilterIndex >= 0) {
      // If the exact same filter exists, remove it (toggle off)
      currentFilters.splice(existingFilterIndex, 1);
    } else {
      // Check if there's already a filter of this type
      const sameTypeIndex = currentFilters.findIndex(f => f.type === filterType);
      if (sameTypeIndex >= 0) {
        // Replace the existing filter of the same type
        currentFilters[sameTypeIndex] = { type: filterType, value: filterValue };
      } else {
        // Add new filter
        currentFilters.push({ type: filterType, value: filterValue });
      }
    }
    
    // Rebuild the URL with all active filters
    const newParams = new URLSearchParams();
    
    // Add all active filters
    currentFilters.forEach(filter => {
      newParams.append('filter', `${filter.type}:${filter.value}`);
    });
    
    // Preserve the current grouping
    const currentGroupBy = params.get('groupBy') || 'shape';
    newParams.set('groupBy', currentGroupBy);
    
    // Update the URL
    const newUrl = new URL(url);
    newUrl.search = newParams.toString();
    
    // Update browser URL without reloading
    window.history.pushState({}, '', newUrl.toString());
    
    // Make the HTMX request with the updated URL
    htmx.ajax('GET', newUrl.toString(), {
      target: '#items-container',
      swap: 'outerHTML',
      headers: { 'HX-Request': 'true' }
    });
  };
});

describe('Filtering Functionality', () => {
  test('should add a single filter and make HTMX request', () => {
    // Call setActiveFilter with a color filter
    setActiveFilter('color', 'red');
    
    // Verify HTMX was called
    expect(htmx.ajax).toHaveBeenCalledTimes(1);
    
    // Get the call arguments
    const [method, url, options] = htmx.ajax.mock.calls[0];
    
    // Verify the request details
    expect(method).toBe('GET');
    // URL encodes : as %3A
    expect(url).toMatch(/filter=color(%3A|:)red/);
    expect(url).toContain('groupBy=shape');
    expect(options.target).toBe('#items-container');
    expect(options.swap).toBe('outerHTML');
    
    // Verify URL was updated in browser history
    expect(window.history.pushState).toHaveBeenCalledTimes(1);
    const pushArgs = window.history.pushState.mock.calls[0];
    expect(pushArgs[0]).toEqual({});
    expect(pushArgs[1]).toBe('');
    expect(pushArgs[2]).toMatch(/filter=color(%3A|:)red/);
  });

  test('should handle multiple filters of different types', () => {
    // Add first filter
    setActiveFilter('color', 'red');
    
    // Update window.location to reflect the first filter
    window.location.search = '?filter=color:red&groupBy=shape';
    window.location.href = 'http://localhost:8080/items?filter=color:red&groupBy=shape';
    
    jest.clearAllMocks();
    
    // Add second filter of different type
    setActiveFilter('shape', 'circle');
    
    // Verify HTMX was called
    expect(htmx.ajax).toHaveBeenCalledTimes(1);
    
    // Get the URL from the call
    const [, url] = htmx.ajax.mock.calls[0];
    
    // Verify both filters are present (URL encoded)
    expect(url).toMatch(/filter=color(%3A|:)red/);
    expect(url).toMatch(/filter=shape(%3A|:)circle/);
    expect(url).toContain('groupBy=shape');
  });

  test('should replace filter when clicking different value of same type', () => {
    // Add first filter
    setActiveFilter('color', 'red');
    
    // Update window.location to reflect the first filter
    window.location.search = '?filter=color:red&groupBy=shape';
    window.location.href = 'http://localhost:8080/items?filter=color:red&groupBy=shape';
    
    jest.clearAllMocks();
    
    // Add different color filter (should replace, not add)
    setActiveFilter('color', 'blue');
    
    // Verify HTMX was called
    expect(htmx.ajax).toHaveBeenCalledTimes(1);
    
    // Get the URL from the call
    const [, url] = htmx.ajax.mock.calls[0];
    
    // Verify only blue filter is present (red was replaced)
    expect(url).toMatch(/filter=color(%3A|:)blue/);
    expect(url).not.toMatch(/filter=color(%3A|:)red/);
    expect(url).toContain('groupBy=shape');
  });

  test('should toggle off filter when clicking same filter again', () => {
    // Add filter
    setActiveFilter('color', 'red');
    
    // Update window.location to reflect the filter
    window.location.search = '?filter=color:red&groupBy=shape';
    window.location.href = 'http://localhost:8080/items?filter=color:red&groupBy=shape';
    
    jest.clearAllMocks();
    
    // Click same filter again (should remove it)
    setActiveFilter('color', 'red');
    
    // Verify HTMX was called
    expect(htmx.ajax).toHaveBeenCalledTimes(1);
    
    // Get the URL from the call
    const [, url] = htmx.ajax.mock.calls[0];
    
    // Verify filter was removed
    expect(url).not.toMatch(/filter=color(%3A|:)red/);
    expect(url).toContain('groupBy=shape');
  });

  test('should preserve groupBy parameter when filtering', () => {
    // Set groupBy to color
    window.location.search = '?groupBy=color';
    window.location.href = 'http://localhost:8080/items?groupBy=color';
    
    // Add filter
    setActiveFilter('shape', 'circle');
    
    // Get the URL from the call
    const [, url] = htmx.ajax.mock.calls[0];
    
    // Verify groupBy was preserved
    expect(url).toContain('groupBy=color');
    expect(url).toMatch(/filter=shape(%3A|:)circle/);
  });

  test('should handle three different filter types simultaneously', () => {
    // Add color filter
    setActiveFilter('color', 'red');
    window.location.search = '?filter=color:red&groupBy=shape';
    window.location.href = 'http://localhost:8080/items?filter=color:red&groupBy=shape';
    jest.clearAllMocks();
    
    // Add shape filter
    setActiveFilter('shape', 'circle');
    window.location.search = '?filter=color:red&filter=shape:circle&groupBy=shape';
    window.location.href = 'http://localhost:8080/items?filter=color:red&filter=shape:circle&groupBy=shape';
    jest.clearAllMocks();
    
    // Add category filter
    setActiveFilter('category', 'A');
    
    // Get the URL from the call
    const [, url] = htmx.ajax.mock.calls[0];
    
    // Verify all three filters are present (URL encoded)
    expect(url).toMatch(/filter=color(%3A|:)red/);
    expect(url).toMatch(/filter=shape(%3A|:)circle/);
    expect(url).toMatch(/filter=category(%3A|:)A/);
    expect(url).toContain('groupBy=shape');
  });

  test('should use outerHTML swap mode', () => {
    // Add filter
    setActiveFilter('color', 'red');
    
    // Get the options from the call
    const [, , options] = htmx.ajax.mock.calls[0];
    
    // Verify swap mode is outerHTML (not innerHTML)
    expect(options.swap).toBe('outerHTML');
  });

  test('should target #items-container', () => {
    // Add filter
    setActiveFilter('color', 'red');
    
    // Get the options from the call
    const [, , options] = htmx.ajax.mock.calls[0];
    
    // Verify target is correct
    expect(options.target).toBe('#items-container');
  });

  test('should include HX-Request header', () => {
    // Add filter
    setActiveFilter('color', 'red');
    
    // Get the options from the call
    const [, , options] = htmx.ajax.mock.calls[0];
    
    // Verify header is present
    expect(options.headers).toEqual({ 'HX-Request': 'true' });
  });

  test('should update browser history with correct URL', () => {
    // Add filter
    setActiveFilter('color', 'red');
    
    // Verify pushState was called
    expect(window.history.pushState).toHaveBeenCalledTimes(1);
    
    // Get the URL from pushState call
    const [, , url] = window.history.pushState.mock.calls[0];
    
    // Verify URL contains filter (URL encoded)
    expect(url).toMatch(/filter=color(%3A|:)red/);
    expect(url).toContain('groupBy=shape');
  });
});

describe('Filter URL Parsing', () => {
  test('should correctly parse multiple filters from URL', () => {
    const params = new URLSearchParams('?filter=color:red&filter=shape:circle&groupBy=shape');
    
    const filters = [];
    for (const filter of params.getAll('filter')) {
      const [type, value] = filter.split(':');
      if (type && value) {
        filters.push({ type, value });
      }
    }
    
    expect(filters).toHaveLength(2);
    expect(filters[0]).toEqual({ type: 'color', value: 'red' });
    expect(filters[1]).toEqual({ type: 'shape', value: 'circle' });
  });

  test('should handle URL with no filters', () => {
    const params = new URLSearchParams('?groupBy=shape');
    
    const filters = [];
    for (const filter of params.getAll('filter')) {
      const [type, value] = filter.split(':');
      if (type && value) {
        filters.push({ type, value });
      }
    }
    
    expect(filters).toHaveLength(0);
  });

  test('should ignore malformed filter parameters', () => {
    const params = new URLSearchParams('?filter=invalid&filter=color:red&groupBy=shape');
    
    const filters = [];
    for (const filter of params.getAll('filter')) {
      const [type, value] = filter.split(':');
      if (type && value) {
        filters.push({ type, value });
      }
    }
    
    expect(filters).toHaveLength(1);
    expect(filters[0]).toEqual({ type: 'color', value: 'red' });
  });
});
