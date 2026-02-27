"use client";
import { useState, useEffect } from "react";
import { Settings, Loader, CheckCircle, XCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { toast } from "sonner";
import { getLLMConfig, saveLLMConfig, testLLMConfig } from "@/api";

export default function SettingsDialog() {
  const [open, setOpen] = useState(false);
  const [baseUrl, setBaseUrl] = useState("");
  const [apiKey, setApiKey] = useState("");
  const [model, setModel] = useState("");
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<"success" | "error" | null>(null);

  useEffect(() => {
    if (open) {
      loadConfig();
    }
  }, [open]);

  const loadConfig = async () => {
    try {
      const config = await getLLMConfig();
      setBaseUrl(config.base_url || "");
      setApiKey(config.api_key || "");
      setModel(config.model || "");
    } catch {
      // Config not available yet
    }
  };

  const handleSave = async () => {
    if (!baseUrl.trim() || !apiKey.trim() || !model.trim()) {
      toast.error("请填写所有字段");
      return;
    }

    setSaving(true);
    try {
      await saveLLMConfig({
        base_url: baseUrl.trim(),
        api_key: apiKey.trim(),
        model: model.trim(),
      });
      toast.success("配置已保存");
      setOpen(false);
    } catch (e) {
      toast.error("保存失败: " + (e as Error).message);
    } finally {
      setSaving(false);
    }
  };

  const handleTest = async () => {
    if (!baseUrl.trim() || !apiKey.trim() || !model.trim()) {
      toast.error("请先填写所有字段");
      return;
    }

    setTesting(true);
    setTestResult(null);
    try {
      const result = await testLLMConfig({
        base_url: baseUrl.trim(),
        api_key: apiKey.trim(),
        model: model.trim(),
      });
      setTestResult("success");
      toast.success("连接成功: " + result);
    } catch (e) {
      setTestResult("error");
      toast.error("连接失败: " + (e as Error).message);
    } finally {
      setTesting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button
          variant="outline"
          size="icon"
          className="fixed top-3 right-24 z-50 rounded-full shadow"
          title="AI 设置"
        >
          <Settings className="h-4 w-4" />
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>AI 设置</DialogTitle>
          <DialogDescription>
            配置 OpenAI 兼容的 API 接口（支持 OpenAI、DeepSeek、通义千问等）
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label htmlFor="base_url">API Base URL</Label>
            <Input
              id="base_url"
              placeholder="https://api.openai.com/v1"
              value={baseUrl}
              onChange={(e) => setBaseUrl(e.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              例如: https://api.openai.com/v1 或 https://api.deepseek.com/v1
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="api_key">API Key</Label>
            <Input
              id="api_key"
              type="password"
              placeholder="sk-..."
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="model">Model</Label>
            <Input
              id="model"
              placeholder="gpt-4o / deepseek-chat / qwen-plus"
              value={model}
              onChange={(e) => setModel(e.target.value)}
            />
          </div>

          {testResult && (
            <div
              className={`flex items-center gap-2 p-2 rounded text-sm ${
                testResult === "success"
                  ? "bg-green-50 text-green-700"
                  : "bg-red-50 text-red-700"
              }`}
            >
              {testResult === "success" ? (
                <CheckCircle className="h-4 w-4" />
              ) : (
                <XCircle className="h-4 w-4" />
              )}
              {testResult === "success" ? "连接测试成功" : "连接测试失败"}
            </div>
          )}
        </div>

        <DialogFooter className="flex-row justify-between sm:justify-between">
          <Button
            variant="outline"
            onClick={handleTest}
            disabled={testing || !baseUrl || !apiKey || !model}
          >
            {testing && <Loader className="h-4 w-4 mr-2 animate-spin" />}
            测试连接
          </Button>
          <div className="flex gap-2">
            <Button variant="outline" onClick={() => setOpen(false)}>
              取消
            </Button>
            <Button onClick={handleSave} disabled={saving}>
              {saving && <Loader className="h-4 w-4 mr-2 animate-spin" />}
              保存
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
