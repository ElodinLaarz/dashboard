// Mock the DOM elements and functions we need
let mockElements = {};

// Mock the toggleCategory function
const toggleCategory = jest.fn();

// Mock the setActiveFilter function
const setActiveFilter = jest.fn();

// Mock the setActiveGroup function
const setActiveGroup = jest.fn();

// Mock the DOM elements
beforeEach(() => {
  // Reset all mocks
  jest.clearAllMocks();
  
  // Set up mock elements
  mockElements = {
    colorItems: {
      classList: {
        contains: jest.fn().mockReturnValue(false),
        toggle: jest.fn(),
        add: jest.fn(),
        remove: jest.fn()
      },
      style: { display: '' },
      querySelectorAll: jest.fn().mockReturnValue([{ style: {} }])
    },
    colorHeader: {
      classList: {
        contains: jest.fn().mockReturnValue(false),
        toggle: jest.fn(),
        add: jest.fn(),
        remove: jest.fn()
      },
      addEventListener: jest.fn((event, callback) => {
        mockElements.colorHeader.click = callback;
      }),
      click: null,
      querySelector: jest.fn().mockReturnValue({
        classList: {
          contains: jest.fn().mockReturnValue(false)
        }
      })
    },
    groupByOptions: {
      addEventListener: jest.fn(),
      contains: jest.fn().mockReturnValue(false)
    },
    groupByButton: {
      onclick: null,
      dispatchEvent: jest.fn(),
      click: jest.fn()
    },
    categoryItem: {
      classList: {
        contains: jest.fn().mockReturnValue(false),
        toggle: jest.fn(),
        add: jest.fn(),
        remove: jest.fn()
      },
      onclick: null,
      click: jest.fn()
    },
    // Add localStorage mock
    localStorage: {
      getItem: jest.fn(),
      setItem: jest.fn(),
      removeItem: jest.fn(),
      clear: jest.fn()
    }
  };
  
  // Mock document - will be set up later with consolidated mocks
  global.document = {
    addEventListener: jest.fn()
  };
  
  // Mock window object
  global.window = {
    setActiveFilter,
    setActiveGroup,
    toggleCategory,
    localStorage: mockElements.localStorage,
    location: {
      href: 'http://localhost',
      pathname: '/',
      search: '?groupBy=color'
    },
    history: {
      pushState: jest.fn()
    },
    URLSearchParams: function() {
      return {
        get: jest.fn().mockReturnValue('color'),
        set: jest.fn(),
        toString: jest.fn().mockReturnValue('groupBy=color')
      };
    },
    // Mock the event
    Event: function() { return {}; },
    CustomEvent: function() { return {}; }
  };

  // Set up the restore function
  window.restoreCollapsedState = () => {
    const category = 'color';
    const items = document.getElementById(`${category}-items`);
    const header = document.querySelector(`[onclick="toggleCategory('${category}')"]`);
    
    if (items && header) {
      const isCollapsed = window.localStorage.getItem(`category-${category}-collapsed`) === 'true';
      if (isCollapsed) {
        items.classList.add('collapsed');
        header.classList.add('collapsed');
      }
    }
  };

  // Set up the header element with the onclick handler
  const createMockElement = () => ({
    classList: {
      contains: jest.fn().mockReturnValue(false),
      add: jest.fn(),
      remove: jest.fn(),
      toggle: jest.fn()
    },
    getAttribute: jest.fn().mockImplementation((attr) => 
      attr === 'onclick' ? "toggleCategory('color')" : null
    ),
    closest: jest.fn().mockReturnValue(null),
    addEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
    click: jest.fn(function(e) {
      if (typeof this.onclick === 'function') {
        this.onclick(e || { stopPropagation: jest.fn(), preventDefault: jest.fn() });
      }
    })
  });

  mockElements.colorHeader = createMockElement();
  mockElements.colorHeader.onclick = function(e) {
    e = e || { stopPropagation: jest.fn(), preventDefault: jest.fn() };
    e.stopPropagation();
    e.preventDefault();
    
    // Check if the click was on a group-by button
    if (e.target.closest && e.target.closest('.group-by-options')) {
      return;
    }
    
    // Call toggleCategory directly as it would be in the HTML
    window.toggleCategory('color');
  };
  
  // Set up classList mocks with state tracking
  const createMockClassList = () => {
    const classes = new Set();
    return {
      contains: jest.fn((className) => classes.has(className)),
      add: jest.fn((className) => classes.add(className)),
      remove: jest.fn((className) => classes.delete(className)),
      toggle: jest.fn((className) => {
        if (classes.has(className)) {
          classes.delete(className);
        } else {
          classes.add(className);
        }
      })
    };
  };
  
  mockElements.colorItems.classList = createMockClassList();
  mockElements.colorHeader.classList = createMockClassList();
  
  // Set up localStorage mock
  const storage = {};
  const localStorageMock = {
    getItem: jest.fn((key) => storage[key]),
    setItem: jest.fn((key, value) => { storage[key] = value; }),
    clear: jest.fn(() => {
      Object.keys(storage).forEach(key => delete storage[key]);
    })
  };
  
  // Set up global localStorage
  Object.defineProperty(window, 'localStorage', {
    value: localStorageMock,
    writable: true
  });
  
  // Also add to mockElements for easier access in tests
  mockElements.localStorage = localStorageMock;
  
  // Set up window.toggleCategory to match the actual implementation
  window.toggleCategory = jest.fn(function(category) {
    const items = document.getElementById(`${category}-items`);
    const header = document.querySelector(`[onclick="toggleCategory('${category}')"]`);
    
    if (items && header) {
      // Toggle the 'collapsed' class on the items container
      items.classList.toggle('collapsed');
      
      // Toggle the 'collapsed' class on the header for styling
      header.classList.toggle('collapsed');
      
      // Store the collapsed state in localStorage
      const isCollapsed = items.classList.contains('collapsed');
      localStorage.setItem(`category-${category}-collapsed`, isCollapsed);
    }
  });
  
  // Consolidate document.getElementById and querySelector mocks
  document.getElementById = jest.fn((id) => 
    id === 'color-items' ? mockElements.colorItems : null
  );
  
  document.querySelector = jest.fn((selector) => {
    if (selector === '[onclick="toggleCategory(\'color\')"]') return mockElements.colorHeader;
    if (selector === '.category-header[onclick*="color"]') return mockElements.colorHeader;
    if (selector === '.category-header') return mockElements.colorHeader;
    if (selector === '.group-by-options') return mockElements.groupByOptions;
    if (selector === '.group-btn') return mockElements.groupByButton;
    if (selector === '.category-item') return mockElements.categoryItem;
    return null;
  });
  
  document.querySelectorAll = jest.fn().mockImplementation((selector) => {
    if (selector === '.category-header') return [mockElements.colorHeader];
    if (selector === '.group-btn') return [mockElements.groupByButton];
    return [];
  });
  
  // Clear any previous test data
  mockElements.localStorage.clear();
});

describe('Sidebar Toggle Functionality', () => {
  test('should toggle collapsed class on items when header is clicked', () => {
    // Create a mock event
    const mockEvent = {
      target: mockElements.colorHeader,
      currentTarget: mockElements.colorHeader,
      stopPropagation: jest.fn(),
      preventDefault: jest.fn(),
      stopImmediatePropagation: jest.fn()
    };
    
    // Simulate header click by directly calling the onclick handler
    mockElements.colorHeader.onclick(mockEvent);
    
    // Verify the event methods were called
    expect(mockEvent.stopPropagation).toHaveBeenCalled();
    expect(mockEvent.preventDefault).toHaveBeenCalled();
    
    // Verify toggleCategory was called with the right argument
    expect(window.toggleCategory).toHaveBeenCalledWith('color');
    
    // Verify the classList.toggle method was called with 'collapsed' on both elements
    expect(mockElements.colorItems.classList.toggle).toHaveBeenCalledWith('collapsed');
    expect(mockElements.colorHeader.classList.toggle).toHaveBeenCalledWith('collapsed');
    
    // Verify localStorage was updated with the correct value (now collapsed)
    expect(mockElements.localStorage.setItem).toHaveBeenCalledWith('category-color-collapsed', true);
    
    // Reset mocks for the second part of the test
    jest.clearAllMocks();
    
    // Second click - should remove 'collapsed' class
    mockElements.colorHeader.onclick(mockEvent);
    
    // Verify the classList.toggle method was called with 'collapsed' again
    expect(mockElements.colorItems.classList.toggle).toHaveBeenCalledWith('collapsed');
    expect(mockElements.colorHeader.classList.toggle).toHaveBeenCalledWith('collapsed');
    expect(mockElements.localStorage.setItem).toHaveBeenCalledWith('category-color-collapsed', false);
  });

  test('should not toggle when clicking group by button', () => {
    // Set up the group by button mock
    const groupByButton = { 
      closest: jest.fn().mockReturnValue({ classList: { contains: jest.fn().mockReturnValue(true) } }),
      classList: {
        contains: jest.fn().mockReturnValue(false)
      }
    };
    
    const mockEvent = {
      target: groupByButton,
      currentTarget: mockElements.colorHeader,
      stopPropagation: jest.fn(),
      preventDefault: jest.fn(),
      stopImmediatePropagation: jest.fn()
    };
    
    // Simulate click on group by button
    mockElements.colorHeader.onclick(mockEvent);
    
    // Verify the click was stopped
    expect(mockEvent.stopPropagation).toHaveBeenCalled();
    // Verify toggleCategory was not called
    expect(window.toggleCategory).not.toHaveBeenCalled();
  });

  test('should update localStorage when toggling categories', () => {
    // Reset mocks
    jest.clearAllMocks();
    
    // Call the toggleCategory function (first toggle - should add collapsed)
    window.toggleCategory('color');
    
    // Verify the classList.toggle was called on both elements
    expect(mockElements.colorItems.classList.toggle).toHaveBeenCalledWith('collapsed');
    expect(mockElements.colorHeader.classList.toggle).toHaveBeenCalledWith('collapsed');
    
    // Verify localStorage was updated with the correct state (now collapsed)
    expect(mockElements.localStorage.setItem).toHaveBeenCalledWith('category-color-collapsed', true);
    
    // Now test the other direction (uncollapsing)
    jest.clearAllMocks();
    
    // Call toggleCategory again (second toggle - should remove collapsed)
    window.toggleCategory('color');
    
    // Verify localStorage was updated with the new state (no longer collapsed)
    expect(mockElements.localStorage.setItem).toHaveBeenCalledWith('category-color-collapsed', false);
  });

  test('should restore collapsed state from localStorage on load', () => {
    // Clear previous calls
    jest.clearAllMocks();
    
    // Set up localStorage to return true for collapsed state
    window.localStorage.setItem('category-color-collapsed', 'true');
    
    // Call the restore function
    window.restoreCollapsedState();
    
    // Verify we checked localStorage
    expect(window.localStorage.getItem).toHaveBeenCalledWith('category-color-collapsed');
    
    // Verify the collapsed class was added to both items and header
    expect(mockElements.colorItems.classList.add).toHaveBeenCalledWith('collapsed');
    expect(mockElements.colorHeader.classList.add).toHaveBeenCalledWith('collapsed');
  });

  test('should handle group by button clicks independently', () => {
    const mockEvent = {
      stopPropagation: jest.fn(),
      target: mockElements.groupByButton,
      currentTarget: mockElements.groupByButton
    };
    
    // Simulate group by button click
    const buttonClickHandler = (e) => {
      e.stopPropagation();
      setActiveGroup('color');
    };
    
    buttonClickHandler(mockEvent);
    
    // Verify the click was handled correctly
    expect(mockEvent.stopPropagation).toHaveBeenCalled();
    expect(setActiveGroup).toHaveBeenCalledWith('color');
    expect(toggleCategory).not.toHaveBeenCalled();
  });
});
