# Aide tui-kanban

## Navigation

| Touche | Action |
|--------|--------|
| `h` / `←` | Colonne précédente |
| `l` / `→` | Colonne suivante |
| `j` / `↓` | Tâche suivante |
| `k` / `↑` | Tâche précédente |
| `q` | Quitter |

## Commandes slash

Appuie sur `/` ou `:` pour ouvrir la barre de commandes.
`Tab` autocomplète, `Esc` ferme.

| Commande | Description |
|----------|-------------|
| `/add <titre>` | Nouvelle tâche dans la colonne active |
| `/edit <id>` | Éditer une tâche |
| `/delete <id>` | Supprimer une tâche (confirmation requise) |
| `/move <id> <colonne>` | Déplacer une tâche |
| `/project <nom>` | Changer de projet |
| `/help` | Afficher ce message |
| `/quit` ou `/q` | Quitter |

## Formulaire de tâche

- `Tab` / `Shift+Tab` : naviguer entre les champs
- `Ctrl+S` : sauvegarder
- `Esc` : annuler

## Fichiers

Les tâches sont stockées dans `~/.kanban/<projet>/TASK-NNN.md`.
La configuration est dans `~/.kanban/config.yaml`.
