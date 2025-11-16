local M = {}

local state = {
  job = nil,
  pending = {},
  seq = 0,
  vault_path = nil,
  cmd = nil,
}

local function log(msg)
  vim.notify("[note-rpc] " .. msg, vim.log.levels.INFO)
end

local function start_daemon(vault_path, rust_cmd)
  if state.job then
    log("daemon already running")
    return
  end

  local cmd = rust_cmd or state.cmd or { "cargo", "run", "-p", "note-daemon" }
  state.cmd = cmd
  state.vault_path = vault_path or state.vault_path or vim.loop.cwd()
  local env = vim.tbl_extend("force", vim.fn.environ(), {
    NOTE_VAULT_PATH = state.vault_path,
  })

  state.job = vim.fn.jobstart(cmd, {
    env = env,
    stdout_buffered = false,
    stderr_buffered = false,
    on_stdout = function(_, data)
      for _, line in ipairs(data) do
        if line and line ~= "" then
          local ok, payload = pcall(vim.json.decode, line)
          if ok and payload.id then
            local cb = state.pending[payload.id]
            state.pending[payload.id] = nil
            if cb then
              cb(payload)
            end
          elseif ok and not payload.id then
            log("notification: " .. vim.inspect(payload))
          elseif line and line ~= "" then
            log("unparseable stdout: " .. line)
          end
        end
      end
    end,
    on_stderr = function(_, data)
      for _, line in ipairs(data) do
        if line and line ~= "" then
          vim.notify("[note-daemon stderr] " .. line, vim.log.levels.WARN)
        end
      end
    end,
    on_exit = function()
      log("note-daemon exited")
      state.job = nil
    end,
  })

  if state.job <= 0 then
    vim.notify("failed to start note-daemon", vim.log.levels.ERROR)
    state.job = nil
  else
    state.vault_path = vault_path
    log("note-daemon started (job " .. state.job .. ")")
  end
end

local function request(method, params, cb)
  if not state.job then
    vim.notify("note-daemon not running; run :NoteDaemonStart", vim.log.levels.ERROR)
    return
  end
  state.seq = state.seq + 1
  local id = state.seq
  if cb then
    state.pending[id] = cb
  end
  local payload_params = params or {}
  if vim.tbl_isempty(payload_params) then
    payload_params = vim.empty_dict()
  end
  local payload = {
    jsonrpc = "2.0",
    id = id,
    method = method,
    params = payload_params,
  }
  vim.fn.chansend(state.job, vim.json.encode(payload) .. "\n")
end

local function render_lines(title, lines)
  local buf = vim.api.nvim_create_buf(false, true)
  vim.api.nvim_buf_set_lines(buf, 0, -1, false, lines)
  vim.api.nvim_buf_set_option(buf, "buftype", "nofile")
  vim.api.nvim_buf_set_option(buf, "bufhidden", "wipe")
  vim.api.nvim_buf_set_option(buf, "modifiable", false)

  local win = vim.api.nvim_open_win(buf, true, {
    relative = "editor",
    width = math.floor(vim.o.columns * 0.6),
    height = math.floor(vim.o.lines * 0.4),
    row = math.floor(vim.o.lines * 0.1),
    col = math.floor(vim.o.columns * 0.2),
    border = "single",
  })
  vim.api.nvim_win_set_option(win, "winhl", "Normal:NormalFloat")
  vim.api.nvim_buf_set_name(buf, title)
end

local function list_tasks()
  request("core.list_tasks", {}, function(resp)
    if resp.error then
      vim.notify("list_tasks error: " .. resp.error.message, vim.log.levels.ERROR)
      return
    end
    local tasks = resp.result or {}
    local lines = { "Tasks (" .. #tasks .. ")" }
    for _, task in ipairs(tasks) do
      local tag_str = ""
      if task.tags and #task.tags > 0 then
        tag_str = " #" .. table.concat(task.tags, " #")
      end
      local status = task.status or "Open"
      local task_id = task.id
      if type(task_id) == "table" and task_id.text then
        task_id = task_id.text
      end
      if type(task_id) ~= "string" then
        task_id = tostring(task_id)
      end
      table.insert(lines, string.format("%s [%s]%s", task_id or "?", status, tag_str))
      table.insert(lines, "  " .. (task.title or ""))
      if type(task.description_md) == "string" then
        table.insert(lines, "    " .. task.description_md:gsub("\n", " "))
      end
      table.insert(lines, "")
    end
    render_lines("note-tasks", lines)
  end)
end

local function open_daily(date)
  request("core.open_daily", { date = date }, function(resp)
    if resp.error then
      vim.notify("open_daily error: " .. resp.error.message, vim.log.levels.ERROR)
      return
    end
    local note = resp.result
    if not (note and note.content) then
      vim.notify("open_daily returned empty note", vim.log.levels.WARN)
      return
    end
    local buf = vim.api.nvim_create_buf(true, false)
    vim.api.nvim_buf_set_lines(buf, 0, -1, false, vim.split(note.content, "\n", { plain = true }))
    vim.api.nvim_buf_set_name(buf, note.path or ("note-" .. date))
    vim.api.nvim_set_current_buf(buf)
  end)
end

local function stop_daemon()
  if state.job and state.job > 0 then
    vim.fn.jobstop(state.job)
    state.job = nil
  end
end

function M.setup(opts)
  opts = opts or {}
  if opts.vault_path then
    state.vault_path = opts.vault_path
  end
  if opts.cmd then
    state.cmd = opts.cmd
  end

  vim.api.nvim_create_user_command("NoteDaemonStart", function(cmd_opts)
    local vault = cmd_opts.args ~= "" and cmd_opts.args or state.vault_path
    start_daemon(vault, state.cmd)
  end, { nargs = "?", complete = "dir" })

  vim.api.nvim_create_user_command("NoteDaemonStop", function()
    stop_daemon()
  end, {})

  vim.api.nvim_create_user_command("NoteListTasks", function()
    list_tasks()
  end, {})

  vim.api.nvim_create_user_command("NoteOpenDaily", function(cmd_opts)
    local date = cmd_opts.args
    if date == "" then
      date = os.date("%Y-%m-%d")
    end
    open_daily(date)
  end, { nargs = "?", complete = function()
    return { os.date("%Y-%m-%d") }
  end })

  if opts.autostart ~= false then
    start_daemon(state.vault_path, state.cmd)
  end
end

return M
