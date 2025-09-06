import { definePreset } from '@primeng/themes';
import Aura from '@primeng/themes/aura';

export const DATACAT_PRIMENG_PRESET = definePreset(Aura, {
  "primitive": {
    "padding": {
      "none": "0",
      "xs": "0.15rem",
      "sm": "0.5rem",
      "md": "1rem",
      "lg": "2rem",
      "xl": "3rem"
    }
  },
  "semantic": {
    "primary": {
      "50": "#fff9f2",
      "100": "#ffe2c2",
      "200": "#ffcc91",
      "300": "#ffb561",
      "400": "#ff9f30",
      "500": "#ff8800",
      "600": "#d97400",
      "700": "#b35f00",
      "800": "#8c4b00",
      "900": "#663600",
      "950": "#402200"
    },
    "colorScheme": {
      "light": {
        "background": "{slate.100}",
        "surface": {
          "50": "#f8fafc",
          "100": "#f1f5f9",
          "200": "#e2e8f0",
          "300": "#cbd5e1",
          "400": "#94a3b8",
          "500": "#64748b",
          "600": "#475569",
          "700": "#334155",
          "800": "#1e293b",
          "900": "#0f172a",
          "950": "#020617"
        },
        "primary": {
          "contrastColor": "#e2e8f0"
        }
      },
      "dark": {
        "background": "{grey.900}",
        "surface": {
          "50": "#fafafa",
          "100": "#f4f4f5",
          "200": "#e4e4e7",
          "300": "#d4d4d8",
          "400": "#a1a1aa",
          "500": "#71717a",
          "600": "#52525b",
          "700": "#3f3f46",
          "800": "#27272a",
          "900": "#18181b",
          "950": "#09090b"
        }
      }
    }
  },
  "components": {
    "tree": {
      "background": "{primary.contrast}",
      "padding": "{padding.sm} {padding.none}",
      "indent": "0rem",
      "node": {
        "padding": "{padding.xs} {padding.none}",
        "gap": "{padding.sm}"
      }
    },
    "splitter": {
      "border": {
        "color": "transparent"
      },
      "background": "{primary.contrast}",
      "gutter": {
        "background": "{primary.contrast}"
      }
    },
    "breadcrumb": {
      "background": "{primary.contrast}",
      "padding": "{padding.none}"
    },
    "tooltip": {
      "border": {
        "radius": "{borderRadius.lg}"
      }
    },
    "panel": {
      "content": {
        "padding": "{padding.none}"
      }
    },
    "inputtext": {
      "padding": {
        "y": "{padding.xs}"
      }
    },
    "datatable": {
      "body": {
        "cell": {
          "padding": "{padding.sm} {padding.md}"
        }
      }
    },
    "fieldset": {
      "padding": "{padding.none} {padding.md}"
    }
  }
});
