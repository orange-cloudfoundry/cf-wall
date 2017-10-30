<div class="panel panel-default">
  <div class="panel-heading" role="tab" id="tgt-{{ .Name }}-h">
    <h4 class="panel-title">
      <a role="button" data-toggle="collapse" data-parent="#tgt" href="#tgt-{{ .Name }}-c" aria-expanded="false" aria-controls="tgt-{{ .Name }}-c">
        {{.Title}}
        <span class="pull-right label label-primary">0</span>
      </a>
    </h4>
  </div>
  <div id="tgt-{{ .Name }}-c" class="panel-collapse collapse" role="tabpanel" aria-labelledby="tgt-{{ .Name }}-h">
    <ul id="tgt-{{ .Name }}" class="list-group">
    </ul>
  </div>
</div>
